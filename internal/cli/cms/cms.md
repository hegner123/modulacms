# CMS Features

Database actions
* create new data type definitions
* create new content Data instances
* read datatype definitions
* read content data instances
* update data type definitions
    * CRUD fields
* update content data instances
    * CRUD fields
* delete data type definitions
* delete data type instances


Creating / Editing a datatype:
* ui areas: page preview, and dialog 
    * use Keys to control actions
        * A - add datatype
        * X - delete datatype
        * Y - redo
        * Z - undo
        * S - save
        * F - show fields for selected datatype
        * I/E - edit highlighted
        * D - duplicate selected datatype
        * C - copy datatype / field
        * P - paste datatype / field
        * J - Navigate up tree
        * K - Navigate down tree
        * H - Navigate up node children
        * L - Navigate down node children
        

    * dialogs:
        * datatype select
        * datatype options
        * field select
        * field input
            * number
            * text
            * validated text
            * text area
            * media picker
        
* create a content data row as Root of specified datatype, create new MODEL struct for assembling content
    * write MODEL struct to db as content data and content field row
* list datatypes approved as children of parent datatype
* list fields of datatype as menu, when cursor matches field index show input for field, after input save to db



