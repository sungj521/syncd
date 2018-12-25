// Copyright 2018 tinystack Author. All Rights Reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
    "errors"
    "fmt"

    "github.com/tinystack/goutil"
    baseModel "github.com/tinystack/syncd/model"
    userModel "github.com/tinystack/syncd/model/user"
)

type User struct {
    ID          int     `json:"id"`
    GroupId     int     `json:"group_id"`
    Name        string  `json:"name"`
    Password    string  `json:"password"`
    Email       string  `json:"email"`
    TrueName    string  `json:"true_name"`
    Mobile      string  `json:"mobile"`
    LockStatus  int     `json:"lock_status"`
}

type UserItem struct {
    ID              int     `json:"id"`
    GroupId         int     `json:"group_id"`
    GroupName       string  `json:"group_name"`
    Name            string  `json:"name"`
    Email           string  `json:"email"`
    LockStatus      int     `json:"lock_status"`
    LastLoginTime   int     `json:"last_login_time"`
    LastLoginIp     string  `json:"last_login_ip"`
}

func (u *User) CreateOrUpdate() error {
    var ok bool
    user := userModel.User{
        ID: u.ID,
        GroupId: u.GroupId,
        Name: u.Name,
        Email: u.Email,
        TrueName: u.TrueName,
        Mobile: u.Mobile,
        LockStatus: u.LockStatus,
    }
    var salt, password string
    if u.Password != "" {
        salt = goutil.StrRandom(10)
        password = goutil.StrMd5(goutil.JoinStrings(u.Password, salt))
    }
    if u.ID > 0 {
        updateData := map[string]interface{}{
            "group_id": user.GroupId,
            "name": user.Name,
            "email": user.Email,
            "true_name": user.TrueName,
            "mobile": user.Mobile,
            "lock_status": user.LockStatus,
        }
        if password != "" {
            updateData["password"] = password
            updateData["salt"] = salt
        }
        ok = userModel.Update(u.ID, updateData)
    } else {
        user.Password = password
        user.Salt = salt
        ok = userModel.Create(&user)
    }
    if !ok {
        return errors.New("user data update failed")
    }
    return nil
}

func (u *User) List(keyword string, offset, limit int) ([]UserItem, int, error) {
    var (
        ok bool
        userId, total int
        where []baseModel.WhereParam
        userList []UserItem
    )
    if keyword != "" {
        var w *baseModel.WhereParam
        switch {
        case goutil.IsInteger(keyword):
            userId = goutil.Str2Int(keyword)
            if userId > 0 {
                w = &baseModel.WhereParam{
                    Field: "id",
                    Prepare: userId,
                }
            }
        case goutil.IsEmail(keyword):
            w = &baseModel.WhereParam{
                Field: "email",
                Tag: "LIKE",
                Prepare: fmt.Sprintf("%%%s%%", keyword),
            }
        case goutil.IsMobile(keyword):
            w = &baseModel.WhereParam{
                Field: "mobile",
                Tag: "LIKE",
                Prepare: fmt.Sprintf("%%%s%%", keyword),
            }
        default:
            w = &baseModel.WhereParam{
                Field: "name",
                Tag: "LIKE",
                Prepare: fmt.Sprintf("%%%s%%", keyword),
            }
        }
        if w != nil {
            where = append(where, *w)
        }
    }
    list, ok := userModel.List(baseModel.QueryParam{
        Fields: "id, name, group_id, email, lock_status, last_login_ip, last_login_time",
        Offset: offset,
        Limit: limit,
        Order: "id DESC",
        Where: where,
    })
    if !ok {
        return nil, 0, errors.New("get user list data failed")
    }
    total, ok = userModel.Total(baseModel.QueryParam{
        Where: where,
    })
    if !ok {
        return nil, 0, errors.New("get user total count failed")
    }
    for _, u := range list {
        userList = append(userList, UserItem{
            ID: u.ID,
            Name: u.Name,
            Email: u.Email,
            GroupId: u.GroupId,
            LockStatus: u.LockStatus,
            LastLoginIp: u.LastLoginIp,
            LastLoginTime: u.LastLoginTime,
        })
    }
    return userList, total, nil
}

func (u *User) Get() error {
    if u.ID == 0 {
        return errors.New("id can not be empty")
    }
    detail, ok := userModel.Get(u.ID)
    if !ok {
        return errors.New("get user detail data failed")
    }

    u.ID = detail.ID
    u.GroupId = detail.GroupId
    u.Name = detail.Name
    u.Password = detail.Password
    u.Email = detail.Email
    u.TrueName = detail.TrueName
    u.Mobile = detail.Mobile
    u.LockStatus = detail.LockStatus

    return nil
}

func (u *User) CheckUserExists() (bool, error) {
    var where []baseModel.WhereParam
    if u.Name != "" {
        where = append(where, baseModel.WhereParam{
            Field: "name",
            Prepare: u.Name,
        })
    }
    if u.Email != "" {
        where = append(where, baseModel.WhereParam{
            Field: "email",
            Prepare: u.Email,
        })
    }
    if u.ID > 0 {
        where = append(where, baseModel.WhereParam{
            Field: "id",
            Tag: "!=",
            Prepare: u.ID,
        })
    }
    detail, ok := userModel.GetOne(baseModel.QueryParam{
        Where: where,
    })
    if !ok {
        return false, errors.New("get user one data failed")
    }
    return detail.ID > 0, nil
}

func (u *User) Delete() error {
    if u.ID == 0 {
        return errors.New("id can not be empty")
    }
    ok := userModel.Delete(u.ID)
    if !ok {
        return errors.New("user delete failed")
    }
    return nil
}

func (u *User) Search() ([]UserItem, error){
    var where []baseModel.WhereParam
    if u.Name != "" {
        where = append(where, baseModel.WhereParam{
            Field: "name",
            Tag: "LIKE",
            Prepare: fmt.Sprintf("%%%s%%", u.Name),
        })
    }
    if u.Email != "" {
        where = append(where, baseModel.WhereParam{
            Field: "email",
            Prepare: u.Email,
        })
    }
    list, ok := userModel.List(baseModel.QueryParam{
        Fields: "id, name, group_id, email, lock_status",
        Order: "id DESC",
        Where: where,
    })
    if !ok {
        return nil, errors.New("get user list failed")
    }

    var userList []UserItem
    for _, l := range list {
        userList = append(userList, UserItem{
            ID: l.ID,
            GroupId: l.GroupId,
            Name: l.Name,
            Email: l.Email,
            LockStatus: l.LockStatus,
        })
    }
    return userList, nil
}

func (u *User) GetListByIds(ids []int) ([]UserItem, error){
    list, ok := userModel.List(baseModel.QueryParam{
        Fields: "id, name, group_id, email, lock_status, last_login_ip, last_login_time",
        Order: "id DESC",
        Where: []baseModel.WhereParam{
            baseModel.WhereParam{
                Field: "id",
                Tag: "IN",
                Prepare: ids,
            },
        },
    })
    if !ok {
        return nil, errors.New("get user list failed")
    }
    var userList []UserItem
    for _, l := range list {
        userList = append(userList, UserItem{
            ID: l.ID,
            GroupId: l.GroupId,
            Name: l.Name,
            Email: l.Email,
            LockStatus: l.LockStatus,
            LastLoginTime: l.LastLoginTime,
            LastLoginIp: l.LastLoginIp,
        })
    }
    return userList, nil
}