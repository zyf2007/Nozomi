package server

import (
	"fmt"

	stderrors "errors"

	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

func tencentErr(err error) string {
	var sdkErr *tcerr.TencentCloudSDKError
	if stderrors.As(err, &sdkErr) {
		return fmt.Sprintf("腾讯云 SES 错误: %s: %s", sdkErr.Code, sdkErr.Message)
	}
	return err.Error()
}
