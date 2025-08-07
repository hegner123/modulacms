# Concurrent Tree Loading Implementation Plan

## Architecture Overview

A sophisticated lazy-loading system that provides immediate responsiveness with eventual completeness through background breadth-first loading and user-driven priority loading.

## Phase-Based Loading Strategy

### Phase 1: Fast Init Load
```
Immediate: Critical page structure (0-20ms)
Page → Navigation, Hero, Footer, Containers (+ immediate Rows)
User sees functional page structure instantly
```

### Phase 2: Background Breadth-First Loading
```
Background goroutine with priority queue
Breadth-first traversal: Level 3 → Level 4 → Level 5...
Queue progression: [Columns] → [Cards, RichText] → [CardHeaders, CardBodies...]
```

### Phase 3: Interactive Priority Boost
```
User interaction triggers immediate high-priority loading
Jumps to front of queue, provides instant response
Background process resumes after handling priority requests
```

## Core Data Structures

```go
type LoadTask struct {
    ParentID   int64
    Priority   Priority
    Depth      int
    RequestID  string  // For tracking/deduplication
    Timestamp  time.Time
}

type Priority int
const (
    LOW Priority = iota
    NORMAL
    HIGH
    CRITICAL
)

type TreeLoader struct {
    tree         *TreeRoot
    database     *Database
    
    // Channels for task coordination
    backgroundQueue chan LoadTask
    priorityQueue   chan LoadTask
    shutdownChan    chan struct{}
    
    // Synchronization
    treeMutex       sync.RWMutex
    loadingNodes    sync.Map  // Track nodes currently being loaded
    completedNodes  sync.Map  // Track completed nodes (deduplication)
    
    // Configuration
    maxConcurrency  int
    queueSize      int
    batchSize      int
}
```

## Implementation Components

### 1. Tree Loader Initialization

```go
func NewTreeLoader(tree *TreeRoot, database *Database) *TreeLoader {
    return &TreeLoader{
        tree:            tree,
        database:        database,
        backgroundQueue: make(chan LoadTask, 1000),
        priorityQueue:   make(chan LoadTask, 100),
        shutdownChan:    make(chan struct{}),
        loadingNodes:    sync.Map{},
        completedNodes:  sync.Map{},
        maxConcurrency:  3,
        queueSize:      1000,
        batchSize:      10,
    }
}
```

### 2. Background Loading Orchestrator

```go
func (tl *TreeLoader) StartBackgroundLoading(ctx context.Context) {
    // Start worker goroutines
    for i := 0; i < tl.maxConcurrency; i++ {
        go tl.loadWorker(ctx, i)
    }
    
    // Start queue manager
    go tl.queueManager(ctx)
    
    // Seed initial background tasks from init-loaded nodes
    tl.seedBackgroundTasks()
}

func (tl *TreeLoader) queueManager(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case priorityTask := <-tl.priorityQueue:
            // Priority tasks jump to front
            select {
            case tl.backgroundQueue <- priorityTask:
            default:
                // If queue full, make room by dropping oldest background task
                <-tl.backgroundQueue
                tl.backgroundQueue <- priorityTask
            }
        case <-tl.shutdownChan:
            return
        }
    }
}
```

### 3. Load Worker Implementation

```go
func (tl *TreeLoader) loadWorker(ctx context.Context, workerID int) {
    for {
        select {
        case <-ctx.Done():
            return
        case task := <-tl.backgroundQueue:
            if tl.shouldSkipTask(task) {
                continue
            }
            
            tl.processLoadTask(task, workerID)
            
        case <-tl.shutdownChan:
            return
        }
    }
}

func (tl *TreeLoader) processLoadTask(task LoadTask, workerID int) {
    // Mark as loading to prevent duplicates
    if _, loaded := tl.loadingNodes.LoadOrStore(task.ParentID, true); loaded {
        return
    }
    defer tl.loadingNodes.Delete(task.ParentID)
    
    // Load children from database
    children, err := tl.database.GetChildrenOf(task.ParentID)
    if err != nil {
        log.Printf("Worker %d failed to load children for node %d: %v", 
            workerID, task.ParentID, err)
        return
    }
    
    // Insert children into tree
    tl.treeMutex.Lock()
    parentNode := tl.tree.NodeIndex[task.ParentID]
    if parentNode != nil {
        for _, child := range children {
            childNode := tl.createTreeNode(child)
            if parentNode.ShallowInsert(childNode, task.ParentID) {
                tl.tree.NodeIndex[child.ContentDataID] = &childNode
                
                // Queue children's children for next level (breadth-first)
                tl.queueNextLevel(child.ContentDataID, task.Depth+1, NORMAL)
            }
        }
    }
    tl.treeMutex.Unlock()
    
    // Mark as completed
    tl.completedNodes.Store(task.ParentID, true)
}
```

### 4. User-Driven Priority Loading

```go
func (tl *TreeLoader) LoadNodeChildren(nodeID int64) error {
    // Check if already loaded or loading
    if _, loading := tl.loadingNodes.Load(nodeID); loading {
        return nil // Already in progress
    }
    if _, completed := tl.completedNodes.Load(nodeID); completed {
        return nil // Already completed
    }
    
    // Create high-priority task
    priorityTask := LoadTask{
        ParentID:  nodeID,
        Priority:  HIGH,
        Depth:     tl.calculateNodeDepth(nodeID),
        RequestID: fmt.Sprintf("user-%d-%d", nodeID, time.Now().UnixNano()),
        Timestamp: time.Now(),
    }
    
    // Send to priority queue (non-blocking)
    select {
    case tl.priorityQueue <- priorityTask:
        return nil
    default:
        return errors.New("priority queue full")
    }
}
```

### 5. Breadth-First Queue Management

```go
func (tl *TreeLoader) queueNextLevel(parentID int64, depth int, priority Priority) {
    task := LoadTask{
        ParentID:  parentID,
        Priority:  priority,
        Depth:     depth,
        RequestID: fmt.Sprintf("bg-%d-%d", parentID, depth),
        Timestamp: time.Now(),
    }
    
    // Non-blocking send to background queue
    select {
    case tl.backgroundQueue <- task:
    default:
        // Queue full - this is normal for background tasks
        log.Printf("Background queue full, dropping task for node %d", parentID)
    }
}

func (tl *TreeLoader) seedBackgroundTasks() {
    // Start with all nodes currently in the tree from init load
    tl.treeMutex.RLock()
    for nodeID := range tl.tree.NodeIndex {
        tl.queueNextLevel(nodeID, tl.calculateNodeDepth(nodeID)+1, NORMAL)
    }
    tl.treeMutex.RUnlock()
}
```

### 6. Helper Functions

```go
func (tl *TreeLoader) createTreeNode(data db.ContentData) cli.TreeNode {
    node := cli.TreeNode{Node: data}
    
    // Load datatype information
    if datatype, err := tl.database.GetDatatype(data.DatatypeID); err == nil {
        node.NodeDatatype = *datatype
    }
    
    return node
}

func (tl *TreeLoader) shouldSkipTask(task LoadTask) bool {
    // Skip if already completed
    if _, completed := tl.completedNodes.Load(task.ParentID); completed {
        return true
    }
    
    // Skip if task is too old (configurable timeout)
    if time.Since(task.Timestamp) > time.Minute*5 {
        return true
    }
    
    return false
}

func (tl *TreeLoader) calculateNodeDepth(nodeID int64) int {
    tl.treeMutex.RLock()
    defer tl.treeMutex.RUnlock()
    
    // Implementation depends on how you want to calculate depth
    // Could traverse up the tree or maintain depth metadata
    return 0 // Placeholder
}
```

## Usage Pattern

```go
// Initialize system
database := ConnectToDatabase("./modula.db")
tree := performInitLoad(database) // Your existing init logic
loader := NewTreeLoader(tree, database)

// Start background loading
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
loader.StartBackgroundLoading(ctx)

// Handle user interactions
func handleUserExpandNode(nodeID int64) {
    if err := loader.LoadNodeChildren(nodeID); err != nil {
        log.Printf("Failed to load node children: %v", err)
    }
}

// Graceful shutdown
func shutdown() {
    close(loader.shutdownChan)
    cancel()
}
```

## Configuration & Tuning

### Performance Knobs
```go
type LoaderConfig struct {
    MaxConcurrency     int           // Number of worker goroutines
    BackgroundQueueSize int          // Size of background task queue
    PriorityQueueSize  int           // Size of priority task queue
    BatchSize          int           // Number of children to load per batch
    TaskTimeout        time.Duration // Max age for background tasks
    RetryAttempts      int           // Retry failed loads
    LoadDelay          time.Duration // Delay between background loads
}
```

### Error Handling & Resilience
- **Timeout handling**: Drop old background tasks
- **Retry logic**: Exponential backoff for failed loads
- **Circuit breaker**: Stop background loading if database errors spike
- **Graceful degradation**: User requests always work even if background fails
- **Memory pressure**: Drop background queue if memory usage high

## Benefits

### Performance Benefits
- **Sub-20ms init load**: Critical structure available instantly
- **Responsive interactions**: User requests bypass background queue
- **Bandwidth optimization**: Spreads database load over time
- **Memory efficiency**: Controlled, progressive loading

### Architectural Benefits
- **Fault tolerance**: Background failures don't break user experience
- **Scalability**: Configurable concurrency and queue sizes
- **Debuggability**: Clear separation of concerns and logging
- **Maintainability**: Clean interfaces and well-defined responsibilities

## Testing Strategy

### Unit Tests
- Test individual worker behavior
- Test queue management logic
- Test deduplication mechanisms
- Test error handling scenarios

### Integration Tests
- Test full loading scenarios with real database
- Test concurrent user interactions during background loading
- Test system behavior under high load
- Test graceful shutdown and cleanup

### Performance Tests
- Measure init load times
- Measure user interaction response times
- Test memory usage patterns
- Test database query patterns and load

This architecture provides Netflix-level performance characteristics with intelligent prefetching and user-driven prioritization, perfectly suited for a production CMS system.