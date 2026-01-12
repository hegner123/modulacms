# VOICE.md - Writing Like a Human, Not an AI

This document contains specific directives to make blog content sound authentically human and avoid AI-generated patterns.

## The Golden Rules

### Rule #1: One Bulleted List Maximum
**Limit: 1 bulleted list per entire article**

Bulleted lists are the biggest tell for AI-generated content. AI loves to organize everything into neat bullet points. Humans don't.

**Instead of:**
```markdown
Benefits of this approach:
- Faster performance
- Better readability
- Easier testing
- Lower memory usage
```

**Write:**
```markdown
This approach is faster, more readable, and easier to test. It also
uses less memory, which matters when you're dealing with large datasets.
```

If you absolutely need a bulleted list for a collection of items, use it once and make it count. Then convert everything else to prose.

### Rule #2: No Numbered Step Lists (Max One Exception)
**Limit: 1 numbered list per article, or zero**

Same problem as bullet points. If you're writing a tutorial with steps, write them as prose with clear transitions.

**Instead of:**
```markdown
1. Install the package
2. Configure the settings
3. Run the build command
```

**Write:**
```markdown
First, install the package. Then you'll need to configure the settings
in your config file. Once that's done, run the build command.
```

### Rule #3: No Emoji, No Symbols
**Never use unless explicitly requested**

No ✨ sparkles, no ✅ checkmarks, no 🚀 rockets, no 💡 lightbulbs. Humans writing technical content don't pepper it with emoji. AI does.

### Rule #4: Humans Don't Use Em Dash (—) for Pauses in Thought
**Never use Em dash (—) to show a pause**

AI loves em dashes to create sophisticated-sounding pauses. Humans don't naturally write that way in casual technical content.

**Instead of:**
```markdown
The hook runs on mount — and that's it.
Performance was terrible — turned out to be the API call.
```

**Write:**
```markdown
The hook runs on mount, and that's it.
Performance was terrible... turned out to be the API call.
```

Use commas, periods, or ellipsis for pauses. Save em dashes for formal writing.

### Rule #5: No Summary Sections
**Skip "In conclusion", "To summarize", "Wrapping up"**

When you're done explaining, just stop. Don't recap what you just said. Readers can scroll up if they forgot something.

End with the last useful thing, not a summary.

**Bad ending:**
```markdown
## Conclusion
In this article, we explored how to optimize React renders by using
memo and useMemo. We learned that premature optimization is bad and
that you should profile first.
```

**Good ending:**
```markdown
So profile first, optimize second. And maybe those re-renders aren'to 
even your problem. They weren't mine - turned out to be a slow API call.
```

## Structural Authenticity

### No Duplicated Talking Points
If you've already explained a concept or made a point, don't repeat it. AI loves to restate the same thing in different words throughout an article. Humans say it once and move on.

Read through your draft. If you find yourself saying the same thing twice, cut one.

### Vary Paragraph Length
AI writes in uniform paragraph blocks. Humans don't.

Mix it up:
- Single sentence paragraphs for emphasis.
- Longer flowing paragraphs when you're building an argument or walking through code.
- Two-sentence paragraphs as transitions.

**Bad (all same length):**
```markdown
React hooks changed how we write components. They let us use state
in functional components. This made code more reusable and cleaner.

Custom hooks take this further. They let us extract component logic.
This makes our code more modular. It also reduces duplication.
```

**Good (varied):**
```markdown
React hooks changed everything.

Instead of class components with lifecycle methods, you write functions
that use state and effects directly. Cleaner, more reusable, less boilerplate.

Custom hooks take this even further. Pull out that duplicated logic
into a hook and share it across components.
```

### Break Grammar Rules (Sparingly)
Start sentences with "And" or "But" when it feels natural. Use sentence fragments. Add conversational pauses with ellipses.

But don't overdo it. Too much and you sound like you're trying too hard.

### Use Asides and Tangents
Add parenthetical thoughts. Brief digressions that add personality without derailing the main point.

```markdown
The useEffect hook runs after every render by default (which, yeah,
is not what you'd expect from the name). You control when it runs
with the dependency array.
```

### Skip Perfect Transitions
You don't need smooth bridges between every section. Sometimes just jump to the next point.

**AI-style transition:**
```markdown
Now that we understand the problem, let's explore potential solutions.
```

**Human-style:**
```markdown
Here's what actually works...
```

## Limit Heading Hierarchies
**Max depth: H2, occasional H3**

AI loves deep heading structures (H1 > H2 > H3 > H4). Most blog posts don't need more than H2, maybe occasional H3.

**Bad:**
```markdown
# Main Title
## Introduction
### Background
#### Historical Context
```

**Good:**
```markdown
# Main Title
## The Problem
## What I Tried
## What Actually Worked
```

## Show the Messy Parts

### Include Failed Attempts
Don't just show the polished final solution. Show what you tried first that didn't work.

```markdown
My first thought was to use a global state manager. Set up Redux,
write the actions, reducers, the whole nine yards. Worked, but felt
like using a sledgehammer to hang a picture frame.

Then I tried Context. Still too much for what I needed.

Turned out a simple custom hook was all it took.
```

### Admit Uncertainty
Say "I don't know" or "I haven't tried X yet" instead of hedging with formal qualifiers.

**AI-style:**
```markdown
While this approach may potentially work in certain scenarios, it
might not be suitable for all use cases and should be evaluated
based on your specific requirements.
```

**Human-style:**
```markdown
This worked for me. Might not work for you if you're dealing with
real-time data - I haven't tried it in that scenario.
```

### Have Actual Opinions
Don't say "both approaches have their merits." Pick a side.

**AI-style:**
```markdown
Both REST and GraphQL are viable options, each with their own strengths
and weaknesses. The choice depends on your use case.
```

**Human-style:**
```markdown
GraphQL is overkill for most projects. REST is simpler, well-understood,
and works fine unless you're actually dealing with complex nested data.
```

## Voice Checklist

Before publishing, check your draft for these AI tells:

- [ ] More than one bulleted list? **Cut it.**
- [ ] Numbered steps? **Convert to prose.**
- [ ] Emoji or symbols? **Remove them.**
- [ ] "In conclusion" or summary section? **Delete it.**
- [ ] All paragraphs the same length? **Vary them.**
- [ ] Perfect grammar throughout? **Break a rule or two.**
- [ ] Smooth transitions everywhere? **Some jumps are fine.**
- [ ] Only showing polished solutions? **Show the failures.**
- [ ] Hedging instead of having opinions? **Pick a side.**
- [ ] Repeating the same point multiple times? **Say it once.**

## The Litmus Test

Read your article out loud. If it sounds like something you'd say to a coworker explaining the problem, you're good.

If it sounds like a textbook or a corporate blog post, rewrite it.
