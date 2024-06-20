package logic

import (
	"encoding/json"
	"errors"

	"github.com/gravitl/netmaker/database"
	"github.com/gravitl/netmaker/models"
)

// Pre-Define Permission Templates for default Roles
var SuperAdminPermissionTemplate = models.UserRolePermissionTemplate{
	ID:         models.SuperAdminRole,
	Default:    true,
	FullAccess: true,
}
var AdminPermissionTemplate = models.UserRolePermissionTemplate{
	ID:         models.AdminRole,
	Default:    true,
	FullAccess: true,
}

var NetworkAdminPermissionTemplate = models.UserRolePermissionTemplate{
	ID:                 models.NetworkAdmin,
	Default:            true,
	IsNetworkRole:      true,
	FullAccess:         true,
	NetworkLevelAccess: make(map[models.RsrcType]map[models.RsrcID]models.RsrcPermissionScope),
}

var NetworkUserPermissionTemplate = models.UserRolePermissionTemplate{
	ID:                  models.NetworkUser,
	Default:             true,
	FullAccess:          false,
	DenyDashboardAccess: false,
	NetworkLevelAccess: map[models.RsrcType]map[models.RsrcID]models.RsrcPermissionScope{
		models.RemoteAccessGwRsrc: {
			models.AllRemoteAccessGwRsrcID: models.RsrcPermissionScope{
				Read: true,
			},
		},
	},
}

func UserRolesInit() {
	d, _ := json.Marshal(SuperAdminPermissionTemplate)
	database.Insert(SuperAdminPermissionTemplate.ID.String(), string(d), database.USER_PERMISSIONS_TABLE_NAME)
	d, _ = json.Marshal(AdminPermissionTemplate)
	database.Insert(AdminPermissionTemplate.ID.String(), string(d), database.USER_PERMISSIONS_TABLE_NAME)
	d, _ = json.Marshal(NetworkAdminPermissionTemplate)
	database.Insert(NetworkAdminPermissionTemplate.ID.String(), string(d), database.USER_PERMISSIONS_TABLE_NAME)
	d, _ = json.Marshal(NetworkUserPermissionTemplate)
	database.Insert(NetworkUserPermissionTemplate.ID.String(), string(d), database.USER_PERMISSIONS_TABLE_NAME)
}

// ListRoles - lists user roles permission templates
func ListRoles() ([]models.UserRolePermissionTemplate, error) {
	data, err := database.FetchRecords(database.USER_PERMISSIONS_TABLE_NAME)
	if err != nil {
		return []models.UserRolePermissionTemplate{}, err
	}
	userRoles := []models.UserRolePermissionTemplate{}
	for _, dataI := range data {
		userRole := models.UserRolePermissionTemplate{}
		err := json.Unmarshal([]byte(dataI), &userRole)
		if err != nil {
			continue
		}
		userRoles = append(userRoles, userRole)
	}
	return userRoles, nil
}

// CreateRole - inserts new role into DB
func CreateRole(r models.UserRolePermissionTemplate) error {
	// check if role already exists
	if r.ID.String() == "" {
		return errors.New("role id cannot be empty")
	}
	_, err := database.FetchRecord(database.USER_PERMISSIONS_TABLE_NAME, r.ID.String())
	if err == nil {
		return errors.New("role already exists")
	}
	d, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return database.Insert(r.ID.String(), string(d), database.USER_PERMISSIONS_TABLE_NAME)
}

// GetRole - fetches role template by id
func GetRole(roleID string) (models.UserRolePermissionTemplate, error) {
	// check if role already exists
	data, err := database.FetchRecord(database.USER_PERMISSIONS_TABLE_NAME, roleID)
	if err != nil {
		return models.UserRolePermissionTemplate{}, errors.New("role already exists")
	}
	ur := models.UserRolePermissionTemplate{}
	err = json.Unmarshal([]byte(data), &ur)
	if err != nil {
		return ur, err
	}
	return ur, nil
}

// UpdateRole - updates role template
func UpdateRole(r models.UserRolePermissionTemplate) error {
	if r.ID.String() == "" {
		return errors.New("role id cannot be empty")
	}
	_, err := database.FetchRecord(database.USER_PERMISSIONS_TABLE_NAME, r.ID.String())
	if err != nil {
		return err
	}
	d, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return database.Insert(r.ID.String(), string(d), database.USER_PERMISSIONS_TABLE_NAME)
}

// DeleteRole - deletes user role
func DeleteRole(rid models.UserRole) error {
	if rid.String() == "" {
		return errors.New("role id cannot be empty")
	}
	users, err := GetUsersDB()
	if err != nil {
		return err
	}
	for _, user := range users {
		for userG := range user.UserGroups {
			ug, err := GetUserGroup(userG)
			if err == nil {
				for _, networkRole := range ug.NetworkRoles {
					if networkRole == rid {
						err = errors.New("role cannot be deleted as active user groups are using this role")
						return err
					}
				}
			}
		}

		if user.PlatformRoleID == rid {
			err = errors.New("active roles cannot be deleted.switch existing users to a new role before deleting")
			return err
		}
		for _, networkRole := range user.NetworkRoles {
			if networkRole == rid {
				err = errors.New("active roles cannot be deleted.switch existing users to a new role before deleting")
				return err
			}
		}
	}
	return database.DeleteRecord(database.USER_PERMISSIONS_TABLE_NAME, rid.String())
}

// CreateUserGroup - creates new user group
func CreateUserGroup(g models.UserGroup) error {
	// check if role already exists
	if g.ID == "" {
		return errors.New("group id cannot be empty")
	}
	_, err := database.FetchRecord(database.USER_GROUPS_TABLE_NAME, g.ID)
	if err == nil {
		return errors.New("group already exists")
	}
	d, err := json.Marshal(g)
	if err != nil {
		return err
	}
	return database.Insert(g.ID, string(d), database.USER_GROUPS_TABLE_NAME)
}

// GetUserGroup - fetches user group
func GetUserGroup(gid models.UserGroupID) (models.UserGroup, error) {
	d, err := database.FetchRecord(database.USER_GROUPS_TABLE_NAME, gid.String())
	if err == nil {
		return models.UserGroup{}, err
	}
	var ug models.UserGroup
	err = json.Unmarshal([]byte(d), &ug)
	if err != nil {
		return ug, err
	}
	return ug, nil
}

// ListUserGroups - lists user groups
func ListUserGroups() ([]models.UserGroup, error) {
	data, err := database.FetchRecords(database.USER_GROUPS_TABLE_NAME)
	if err != nil {
		return []models.UserGroup{}, err
	}
	userGroups := []models.UserGroup{}
	for _, dataI := range data {
		userGroup := models.UserGroup{}
		err := json.Unmarshal([]byte(dataI), &userGroup)
		if err != nil {
			continue
		}
		userGroups = append(userGroups, userGroup)
	}
	return userGroups, nil
}

// UpdateUserGroup - updates new user group
func UpdateUserGroup(g models.UserGroup) error {
	// check if group exists
	if g.ID == "" {
		return errors.New("group id cannot be empty")
	}
	_, err := database.FetchRecord(database.USER_GROUPS_TABLE_NAME, g.ID)
	if err != nil {
		return err
	}
	d, err := json.Marshal(g)
	if err != nil {
		return err
	}
	return database.Insert(g.ID, string(d), database.USER_GROUPS_TABLE_NAME)
}

// DeleteUserGroup - deletes user group
func DeleteUserGroup(gid models.UserGroupID) error {
	users, err := GetUsersDB()
	if err != nil {
		return err
	}
	for _, user := range users {
		delete(user.UserGroups, gid)
		UpsertUser(user)
	}
	return database.DeleteRecord(database.USER_GROUPS_TABLE_NAME, gid.String())
}

func HasNetworkRsrcScope(permissionTemplate models.UserRolePermissionTemplate, netid string, rsrcType models.RsrcType, rsrcID models.RsrcID, op string) bool {
	if permissionTemplate.FullAccess {
		return true
	}

	rsrcScope, ok := permissionTemplate.NetworkLevelAccess[rsrcType]
	if !ok {
		return false
	}
	_, ok = rsrcScope[rsrcID]
	return ok
}
