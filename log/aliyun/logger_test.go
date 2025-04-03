/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2022/9/9 12:03 下午
 * @Desc: TODO
 */

package aliyun_test

import (
	"github.com/develop-top/due/log/aliyun/v2"
	"testing"
)

var logger = aliyun.NewLogger()

func TestNewLogger(t *testing.T) {
	defer logger.Close()

	logger.Info("info")
}
