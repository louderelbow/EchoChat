package dto

// ChatMessageRequest WebSocket 上行消息结构（与原 ChatMessageRequest 完全一致）
type ChatMessageRequest struct {
	SessionId  string `json:"session_id"`
	Type       int8   `json:"type"`
	Content    string `json:"content"`
	Url        string `json:"url"`
	SendId     string `json:"send_id"`
	SendName   string `json:"send_name"`
	SendAvatar string `json:"send_avatar"`
	ReceiveId  string `json:"receive_id"`
	FileSize   string `json:"file_size"`
	FileType   string `json:"file_type"`
	FileName   string `json:"file_name"`
	AVdata     string `json:"av_data"`
}

// AVData 音视频信令
type AVData struct {
	MessageId string `json:"messageId"`
	Type      string `json:"type"`
}

// GetMessageListRespond 单聊消息（亦是 WS 下发结构）
type GetMessageListRespond struct {
	SendId     string `json:"send_id"`
	SendName   string `json:"send_name"`
	SendAvatar string `json:"send_avatar"`
	ReceiveId  string `json:"receive_id"`
	Type       int8   `json:"type"`
	Content    string `json:"content"`
	Url        string `json:"url"`
	FileType   string `json:"file_type"`
	FileName   string `json:"file_name"`
	FileSize   string `json:"file_size"`
	CreatedAt  string `json:"created_at"`
}

// GetGroupMessageListRespond 群聊消息（亦是 WS 下发结构）
type GetGroupMessageListRespond struct {
	SendId     string `json:"send_id"`
	SendName   string `json:"send_name"`
	SendAvatar string `json:"send_avatar"`
	ReceiveId  string `json:"receive_id"`
	Type       int8   `json:"type"`
	Content    string `json:"content"`
	Url        string `json:"url"`
	FileType   string `json:"file_type"`
	FileName   string `json:"file_name"`
	FileSize   string `json:"file_size"`
	CreatedAt  string `json:"created_at"`
}

// AVMessageRespond 音视频消息下发结构
type AVMessageRespond struct {
	SendId     string `json:"send_id"`
	SendName   string `json:"send_name"`
	SendAvatar string `json:"send_avatar"`
	ReceiveId  string `json:"receive_id"`
	Type       int8   `json:"type"`
	Content    string `json:"content"`
	Url        string `json:"url"`
	FileType   string `json:"file_type"`
	FileName   string `json:"file_name"`
	FileSize   string `json:"file_size"`
	CreatedAt  string `json:"created_at"`
	AVdata     string `json:"av_data"`
}
