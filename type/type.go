package Type

const (
	// 数据类型
	STRING       int32 = 0
	INT32        int32 = 1
	INT64        int32 = 2
	STRING_ARRAY int32 = 3
	INT32_ARRAY  int32 = 4
	INT64_ARRAY  int32 = 5
	STRING_SET   int32 = 6
	INT32_SET    int32 = 7
	INT64_SET    int32 = 8
	JSON         int32 = 9

	//操作类型
	NEW_USER int32 = 101
	DEL_USER int32 = 102
	DEL_KEY  int32 = 103

	NEW_CHAN  int32 = 104
	DEL_CHAN  int32 = 105
	JOIN_CHAN int32 = 106
	EXIT_CHAN int32 = 107
)

// 修改密码可以看作: 先删除用户，然后添加用户
