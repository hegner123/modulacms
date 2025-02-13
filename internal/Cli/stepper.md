# Stepper

## Cms Steps

## Database Steps
1. Table
    * select table
2. Method
    * Create
    * Read
    * Update
    * Delete
3. Method UI
    * Create
        * Input mapping to map columns with input struct including key, type, and value
        * Enter to confirm Insert, dialog to create more, choose new method, return to main menu, or quit
        * Q to quit. Left arrow or backspace to return to method selection
    * Read
        * Call list db function and map results to tea table?
        * Q to quit. Left arrow or backspace to return to method selection
    * Update
        * Call list db function and map results to tea table with cursor for selection
        * Input ui with values replaced with stored values
        * on Q in editor save and return to update table
    * Delete
        * Call list db function and map results to tea table with cursor for selection
        * Selection ui with delete multiple rows, call to batch delete commands with id
        * on Q in editor save and return to update table
