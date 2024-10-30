package deviceshifulwm2m

import (
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/stretchr/testify/assert"
)

func TestCreateLwM2MInstructions(t *testing.T) {
	// 初始化测试数据
	dsInstructions := &deviceshifubase.DeviceShifuInstructions{
		Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{
			"instruction1": {
				DeviceShifuProtocolProperties: map[string]string{
					objectIdStr:      "123",
					enableObserveStr: "true",
				},
			},
			"instruction2": {
				DeviceShifuProtocolProperties: map[string]string{
					objectIdStr:      "456",
					enableObserveStr: "false",
				},
			},
		},
	}

	// 调用待测试函数
	result := CreateLwM2MInstructions(dsInstructions)

	// 断言结果不为空
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Instructions))

	// 检查每个指令的属性是否正确转换
	instruction1 := result.Instructions["instruction1"]
	assert.NotNil(t, instruction1)
	assert.Equal(t, "123", instruction1.ObjectId)
	assert.True(t, instruction1.EnableObserve)

	instruction2 := result.Instructions["instruction2"]
	assert.NotNil(t, instruction2)
	assert.Equal(t, "456", instruction2.ObjectId)
	assert.False(t, instruction2.EnableObserve)
}

func TestCreateLwM2MInstructions_EmptyInstructions(t *testing.T) {
	// 测试空的指令映射
	dsInstructions := &deviceshifubase.DeviceShifuInstructions{
		Instructions: map[string]*deviceshifubase.DeviceShifuInstruction{},
	}

	// 调用待测试函数
	result := CreateLwM2MInstructions(dsInstructions)

	// 断言结果不为空，且指令映射为空
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Instructions))
}
