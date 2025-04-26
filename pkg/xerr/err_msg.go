package xerr

var codeText = map[int]string{
	SERVER_COMMON_ERROR: "Serivce Error, Please try again later",
	REQUEST_PARAM_ERROR: "Request param error",
	TOKEN_EXPIRE_ERROR:  "Token expired, please login again",
	DB_ERROR:            "Database busy, please try again later",
}

func ErrMsg(errcode int) string {
	if msg, ok := codeText[errcode]; ok {
		return msg
	}
	return codeText[SERVER_COMMON_ERROR]
}
