const CreateFieldComponent = require("./create-field-component")
const LoginForm = require("./login")
const MediaPickerComponent = require("./mediapicker")
const TextFieldComponent = require("./textfield")
const UserManagementComponent = require("./admin/user-management")
const MediaUploadComponent = require ("./mediaupload")


if (CreateFieldComponent === undefined){
    console.error("CreateFieldComponent is undefined")
    throw new Error("CreateFieldComponent is undefined")
}
if (LoginForm === undefined){
    console.error("LoginForm is undefined")
    throw new Error("LoginForm is undefined")
}
if (MediaPickerComponent === undefined){
    console.error("MediaPickerComponent is undefined")
    throw new Error("MediaPickerComponent is undefined")
}
if (TextFieldComponent === undefined){
    console.error("TextFieldComponent is undefined")
    throw new Error("TextFieldComponent is undefined")
}
if (UserManagementComponent === undefined){
    console.error("UserManagementComponent is undefined")
    throw new Error("UserManagementComponent is undefined")
}
if (MediaUploadComponent === undefined){
    console.error("MediaUploadComponent is undefined")
    throw new Error("MediaUploadComponent is undefined")
}
