package enum

// 用户状态（对应原 user_status_enum）
const (
	UserNormal  int8 = 0
	UserDisable int8 = 1
)

// 消息类型（对应原 message_type_enum）
const (
	MsgText         int8 = 0
	MsgFile         int8 = 2
	MsgAudioOrVideo int8 = 3
)

// 消息发送状态（对应原 message_status_enum）
const (
	MsgUnsent int8 = 0
)

// 联系类型（对应原 contact_type_enum）
const (
	ContactUser  int8 = 0
	ContactGroup int8 = 1
)

// 联系状态（对应原 contact_status_enum）
const (
	ContactNormal       int8 = 0
	ContactBeBlack      int8 = 1
	ContactBlack        int8 = 2
	ContactBeDelete     int8 = 3
	ContactDelete       int8 = 4
	ContactQuitGroup    int8 = 6
	ContactKickOutGroup int8 = 7
)

// 申请状态（对应原 contact_apply_status_enum）
const (
	ApplyPending int8 = 0
	ApplyAgree   int8 = 1
	ApplyRefuse  int8 = 2
	ApplyBlack   int8 = 3
)

// 群状态（对应原 group_status_enum）
const (
	GroupNormal  int8 = 0
	GroupDisable int8 = 1
)
