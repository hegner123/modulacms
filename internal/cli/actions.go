package cli

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)


type dbErrMsg struct {
	Error error
}

// TODO Add default case for generic operations
func (m *Model) CLICreate(c *config.Config,table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*c)
		con, _, err := d.GetConnection()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return dbErrMsg{Error: err}
		}
		valuesMap := make(map[string]string, 0)
		for i, v := range m.formValues {
			valuesMap[m.headers[i]] = *v
		}

		defer func() tea.Msg {
			if closeErr := con.Close(); closeErr != nil && err == nil {
				err = closeErr
				return dbErrMsg{Error: err}
			}
			return nil
		}()
		jsonData, err := json.Marshal(valuesMap)
		if err != nil {
			return dbErrMsg{Error: err}
		}
		m.formValues = make([]*string, 0)
		switch table {
		case db.Admin_content_data:
			var result db.CreateAdminContentDataFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateAdminContentDataParams(result)
			utility.DefaultLogger.Finfo("", params)
			d.CreateAdminContentData(params)
		case db.Admin_content_fields:
			var result db.CreateAdminContentFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateAdminContentFieldParams(result)
			d.CreateAdminContentField(params)
		case db.Admin_datatype:
			var result db.CreateAdminDatatypeFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateAdminDatatypeParams(result)
			d.CreateAdminDatatype(params)
		case db.Admin_datatype_fields:
			var result db.CreateAdminDatatypeFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateAdminDatatypeFieldParams(result)
			d.CreateAdminDatatypeField(params)
		case db.Admin_field:
			var result db.CreateAdminFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateAdminFieldParams(result)
			d.CreateAdminField(params)
		case db.Admin_route:
			var result db.CreateAdminRouteFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateAdminRouteParams(result)
			d.CreateAdminRoute(params)
		case db.Content_data:
			var result db.CreateContentDataFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateContentDataParams(result)
			d.CreateContentData(params)
		case db.Content_fields:
			var result db.CreateContentFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateContentFieldParams(result)
			d.CreateContentField(params)
		case db.Datatype:
			var result db.CreateDatatypeFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateDatatypeParams(result)
			d.CreateDatatype(params)
		case db.Datatype_fields:
			var result db.CreateDatatypeFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateDatatypeFieldParams(result)
			d.CreateDatatypeField(params)
		case db.Field:
			var result db.CreateFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateFieldParams(result)
			d.CreateField(params)
		case db.MediaT:
			var result db.CreateMediaFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateMediaParams(result)
			d.CreateMedia(params)
		case db.Media_dimension:
			var result db.CreateMediaDimensionFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateMediaDimensionParams(result)
			d.CreateMediaDimension(params)
		case db.Role:
			var result db.CreateRoleFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateRoleParams(result)
			d.CreateRole(params)
		case db.Route:
			var result db.CreateRouteFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateRouteParams(result)
			d.CreateRoute(params)
		case db.Session:
			var result db.CreateSessionFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateSessionParams(result)
			_, err := d.CreateSession(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Table:
			var result struct{ Label string }
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			d.CreateTable(result.Label)
		case db.Token:
			var result db.CreateTokenFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateTokenParams(result)
			d.CreateToken(params)
		case db.User:
			var result db.CreateUserFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateUserParams(result)
			_, err := d.CreateUser(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.User_oauth:
			var result db.CreateUserOauthFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapCreateUserOauthParams(result)
			_, err := d.CreateUserOauth(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		}

		return nil
	}
}

// TODO Add default case for generic operations
func (m *Model) CLIUpdate(c *config.Config, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*c)
		con, _, err := d.GetConnection()
		if err != nil {
			return dbErrMsg{Error: err}
		}
		valuesMap := make(map[string]string)
		for i, v := range m.formValues {
			valuesMap[m.headers[i]] = *v
		}

		defer func() tea.Msg {
			if closeErr := con.Close(); closeErr != nil && err == nil {
				return dbErrMsg{Error: err}
			}
			return nil
		}()
		jsonData, err := json.Marshal(valuesMap)
		m.formValues = make([]*string, 0)
		if err != nil {
			return dbErrMsg{Error: err}
		}
		switch table {
		case db.Admin_content_data:
			var result db.UpdateAdminContentDataFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}

			params := db.MapUpdateAdminContentDataParams(result)
			_, err := d.UpdateAdminContentData(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Admin_content_fields:
			var result db.UpdateAdminContentFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateAdminContentFieldParams(result)
			_, err := d.UpdateAdminContentField(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Admin_datatype:
			var result db.UpdateAdminDatatypeFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			utility.DefaultLogger.Finfo("", result)
			params := db.MapUpdateAdminDatatypeParams(result)
			_, err := d.UpdateAdminDatatype(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Admin_datatype_fields:
			var result db.UpdateAdminDatatypeFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			utility.DefaultLogger.Finfo("", result)
			params := db.MapUpdateAdminDatatypeFieldParams(result)
			_, err := d.UpdateAdminDatatypeField(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Admin_field:
			var result db.UpdateAdminFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateAdminFieldParams(result)
			_, err := d.UpdateAdminField(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Admin_route:
			var result db.UpdateAdminRouteFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateAdminRouteParams(result)
			_, err := d.UpdateAdminRoute(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Content_data:
			var result db.UpdateContentDataFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateContentDataParams(result)
			_, err := d.UpdateContentData(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Content_fields:
			var result db.UpdateContentFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateContentFieldParams(result)
			_, err := d.UpdateContentField(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Datatype:
			var result db.UpdateDatatypeFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateDatatypeParams(result)
			_, err := d.UpdateDatatype(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Datatype_fields:
			var result db.UpdateDatatypeFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateDatatypeFieldParams(result)
			_, err := d.UpdateDatatypeField(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Field:
			var result db.UpdateFieldFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateFieldParams(result)
			_, err := d.UpdateField(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.MediaT:
			var result db.UpdateMediaFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateMediaParams(result)
			_, err := d.UpdateMedia(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Media_dimension:
			var result db.UpdateMediaDimensionFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateMediaDimensionParams(result)
			_, err := d.UpdateMediaDimension(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Role:
			var result db.UpdateRoleFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateRoleParams(result)
			_, err := d.UpdateRole(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Route:
			var result db.UpdateRouteFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateRouteParams(result)
			_, err := d.UpdateRoute(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Session:
			var result db.UpdateSessionFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateSessionParams(result)
			_, err := d.UpdateSession(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Table:
			var result db.UpdateTableFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateTableParams(result)
			_, err := d.UpdateTable(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Token:
			var result db.UpdateTokenFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateTokenParams(result)
			_, err := d.UpdateToken(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.User:
			var result db.UpdateUserFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateUserParams(result)
			_, err := d.UpdateUser(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.User_oauth:
			var result db.UpdateUserOauthFormParams
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return dbErrMsg{Error: err}
			}
			params := db.MapUpdateUserOauthParams(result)
			_, err := d.UpdateUserOauth(params)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		}

		return nil
	}
}

// TODO Add default case for generic operations
func (m Model) CLIDelete(c *config.Config,table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		d := db.ConfigDB(*c)
		con, _, err := d.GetConnection()
		if err != nil {
			return dbErrMsg{Error: err}
		}
		defer func() tea.Msg {
			if closeErr := con.Close(); closeErr != nil && err == nil {
				return dbErrMsg{Error: err}
			}
			return nil
		}()
		s := make(map[string]string, 0)
		utility.DefaultLogger.Fdebug("row", m.rows[m.cursor][0])
		s["ID"] = m.rows[m.cursor][0]
		switch table {
		case db.Admin_content_data:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteAdminContentData(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Admin_content_fields:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteAdminContentField(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Admin_datatype:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteAdminDatatype(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Admin_datatype_fields:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteAdminDatatypeField(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Admin_field:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteAdminField(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Admin_route:
			var result struct{ ID int64 }
			err := d.DeleteAdminRoute(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Content_data:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteContentData(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Content_fields:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteContentField(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Datatype:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteDatatype(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Datatype_fields:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteDatatypeField(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Field:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteField(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.MediaT:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteMedia(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Media_dimension:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteMediaDimension(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Role:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteRole(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Route:
			var result struct{ ID int64 }
			err := d.DeleteRoute(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Session:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteSession(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Table:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteTable(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.Token:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteToken(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.User:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteUser(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		case db.User_oauth:
			var result struct{ ID int64 }
			result.ID = m.GetIDRow()
			err := d.DeleteUserOauth(result.ID)
			if err != nil {
				return dbErrMsg{Error: err}
			}
		}

		return nil

	}
}
