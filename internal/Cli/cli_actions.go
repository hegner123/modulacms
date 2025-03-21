package cli

import (
	"encoding/json"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
)

func (m model) CLICreate(table db.DBTable) error {
	logFile, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer logFile.Close()
	d := db.ConfigDB(config.Env)
	con, _, err := d.GetConnection()
	if err != nil {
		return err
	}
	fmt.Fprintln(logFile, m.formValues)
	for i, v := range m.formValues {
		if v == nil {
			continue
		}
		headers := m.headers
		header := headers[i]
		m.formMap[header] = *v
	}
	fmt.Fprintln(logFile, m.formMap)
	defer con.Close()
	jsonData, err := json.Marshal(m.formMap)
	if err != nil {
		ErrLog.Fatal("", err)
	}
	switch table {
	case db.Admin_content_data:
		var result db.CreateAdminContentDataFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateAdminContentDataParams(result)
		d.CreateAdminContentData(params)
	case db.Admin_content_fields:
		var result db.CreateAdminContentFieldFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateAdminContentFieldParams(result)
		d.CreateAdminContentField(params)
	case db.Admin_datatype:
		var result db.CreateAdminDatatypeFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateAdminDatatypeParams(result)
		d.CreateAdminDatatype(params)
	case db.Admin_field:
		var result db.CreateAdminFieldFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateAdminFieldParams(result)
		d.CreateAdminField(params)
	case db.Admin_route:
		var result db.CreateAdminRouteFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateAdminRouteParams(result)
		d.CreateAdminRoute(params)
	case db.Content_data:
		var result db.CreateContentDataFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateContentDataParams(result)
		d.CreateContentData(params)
	case db.Content_fields:
		var result db.CreateContentDataFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateContentDataParams(result)
		d.CreateContentData(params)
	case db.Datatype:
		var result db.CreateContentFieldFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateContentFieldParams(result)
		d.CreateContentField(params)
	case db.Field:
		var result db.CreateFieldFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateFieldParams(result)
		d.CreateField(params)
	case db.MediaT:
		var result db.CreateMediaFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateMediaParams(result)
		d.CreateMedia(params)
	case db.Media_dimension:
		var result db.CreateMediaDimensionFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateMediaDimensionParams(result)
		d.CreateMediaDimension(params)
	case db.Role:
		var result db.CreateRoleFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateRoleParams(result)
		d.CreateRole(params)
	case db.Route:
		var result db.CreateRouteFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateRouteParams(result)
		d.CreateRoute(params)
	case db.Session:
		var result db.CreateSessionFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateSessionParams(result)
		_, err := d.CreateSession(params)
		if err != nil {
			return err
		}
	case db.Table:
		var result struct{ Label string }
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		d.CreateTable(result.Label)
	case db.Token:
		var result db.CreateTokenFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateTokenParams(result)
		d.CreateToken(params)
	case db.User:
		var result db.CreateUserFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateUserParams(result)
		_, err := d.CreateUser(params)
		if err != nil {
			return err
		}
	case db.User_oauth:
		var result db.CreateUserOauthFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapCreateUserOauthParams(result)
		_, err := d.CreateUserOauth(params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m model) CLIUpdate(table db.DBTable) error {
	logFile, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer logFile.Close()
	d := db.ConfigDB(config.Env)
	con, _, err := d.GetConnection()
	if err != nil {
		return err
	}
	fmt.Fprintln(logFile, m.formValues)
	for i, v := range m.formValues {
		if v == nil {
			continue
		}
		headers := m.headers
		header := headers[i]
		m.formMap[header] = *v
	}
	fmt.Fprintln(logFile, m.formMap)
	defer con.Close()
	jsonData, err := json.Marshal(m.formMap)
	if err != nil {
		ErrLog.Fatal("", err)
	}
	switch table {
	case db.Admin_content_data:
		var result db.UpdateAdminContentDataFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateAdminContentDataParams(result)
		_, err := d.UpdateAdminContentData(params)
		if err != nil {
			return err
		}
	case db.Admin_content_fields:
		var result db.UpdateAdminContentFieldFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateAdminContentFieldParams(result)
		_, err := d.UpdateAdminContentField(params)
		if err != nil {
			return err
		}
	case db.Admin_datatype:
		var result db.UpdateAdminDatatypeFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateAdminDatatypeParams(result)
		_, err := d.UpdateAdminDatatype(params)
		if err != nil {
			return err
		}
	case db.Admin_field:
		var result db.UpdateAdminFieldFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateAdminFieldParams(result)
		_, err := d.UpdateAdminField(params)
		if err != nil {
			return err
		}
	case db.Admin_route:
		var result db.UpdateAdminRouteFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateAdminRouteParams(result)
		_, err := d.UpdateAdminRoute(params)
		if err != nil {
			return err
		}
	case db.Content_data:
		var result db.UpdateContentDataFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateContentDataParams(result)
		_, err := d.UpdateContentData(params)
		if err != nil {
			return err
		}
	case db.Content_fields:
		var result db.UpdateContentDataFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateContentDataParams(result)
		_, err := d.UpdateContentData(params)
		if err != nil {
			return err
		}
	case db.Datatype:
		var result db.UpdateContentFieldFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateContentFieldParams(result)
		_, err := d.UpdateContentField(params)
		if err != nil {
			return err
		}
	case db.Field:
		var result db.UpdateFieldFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateFieldParams(result)
		_, err := d.UpdateField(params)
		if err != nil {
			return err
		}
	case db.MediaT:
		var result db.UpdateMediaFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateMediaParams(result)
		_, err := d.UpdateMedia(params)
		if err != nil {
			return err
		}
	case db.Media_dimension:
		var result db.UpdateMediaDimensionFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateMediaDimensionParams(result)
		_, err := d.UpdateMediaDimension(params)
		if err != nil {
			return err
		}
	case db.Role:
		var result db.UpdateRoleFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateRoleParams(result)
		_, err := d.UpdateRole(params)
		if err != nil {
			return err
		}
	case db.Route:
		var result db.UpdateRouteFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateRouteParams(result)
		_, err := d.UpdateRoute(params)
		if err != nil {
			return err
		}
	case db.Session:
		var result db.UpdateSessionFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateSessionParams(result)
		_, err := d.UpdateSession(params)
		if err != nil {
			return err
		}
	case db.Table:
		var result db.UpdateTableFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateTableParams(result)
		_, err := d.UpdateTable(params)
		if err != nil {
			return err
		}
	case db.Token:
		var result db.UpdateTokenFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateTokenParams(result)
		_, err := d.UpdateToken(params)
		if err != nil {
			return err
		}
	case db.User:
		var result db.UpdateUserFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateUserParams(result)
		_, err := d.UpdateUser(params)
		if err != nil {
			return err
		}
	case db.User_oauth:
		var result db.UpdateUserOauthFormParams
		if err := json.Unmarshal(jsonData, &result); err != nil {
			ErrLog.Fatal("", err)
		}
		params := db.MapUpdateUserOauthParams(result)
		_, err := d.UpdateUserOauth(params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m model) CLIDelete(table db.DBTable) error {
	d := db.ConfigDB(config.Env)
	con, _, err := d.GetConnection()
	if err != nil {
		return err
	}
	defer con.Close()
	jsonData, err := json.Marshal(m.formMap)
	if err != nil {
		return err
	}
	switch table {
	case db.Admin_content_data:
		var result struct{ ID int64 }
        result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteAdminContentData(result.ID)
		if err != nil {
			return err
		}
	case db.Admin_content_fields:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteAdminContentField(result.ID)
		if err != nil {
			return err
		}
	case db.Admin_datatype:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteAdminDatatype(result.ID)
		if err != nil {
			return err
		}
	case db.Admin_field:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteAdminField(result.ID)
		if err != nil {
			return err
		}
	case db.Admin_route:
		var result struct{ ID string }
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteAdminRoute(result.ID)
		if err != nil {
			return err
		}
	case db.Content_data:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteContentData(result.ID)
		if err != nil {
			return err
		}
	case db.Content_fields:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteContentField(result.ID)
		if err != nil {
			return err
		}
	case db.Datatype:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteDatatype(result.ID)
		if err != nil {
			return err
		}
	case db.Field:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteField(result.ID)
		if err != nil {
			return err
		}
	case db.MediaT:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteMedia(result.ID)
		if err != nil {
			return err
		}
	case db.Media_dimension:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteMediaDimension(result.ID)
		if err != nil {
			return err
		}
	case db.Role:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteRole(result.ID)
		if err != nil {
			return err
		}
	case db.Route:
		var result struct{ ID string }
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteRoute(result.ID)
		if err != nil {
			return err
		}
	case db.Session:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteSession(result.ID)
		if err != nil {
			return err
		}
	case db.Table:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteTable(result.ID)
		if err != nil {
			return err
		}
	case db.Token:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteToken(result.ID)
		if err != nil {
			return err
		}
	case db.User:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteUser(result.ID)
		if err != nil {
			return err
		}
	case db.User_oauth:
		var result struct{ ID int64 }
		result.ID = m.GetIDRow()
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return err
		}
		err := d.DeleteUserOauth(result.ID)
		if err != nil {
			return err
		}
	}

	return nil

}
