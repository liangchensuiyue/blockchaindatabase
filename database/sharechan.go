package database

import (
	"errors"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
)

/*
newchan
	key: channame_username
	value: 加入该chan的密钥
	datatype: newchan
创建成功之后会生成一条加入chan的密钥,密钥结构:username.加入密钥.chan内数据加密密钥

joinchan
	key: channame_username
	value:username 密钥
	datatype: add_chan

delchan
	key: channame_username
	value:用户密码hash

	ps: 先校验用户，然后在添加操作
exitchan
	key:channame_username
	value:用户密码hash 用户名称
*/sh 用户名称
*/
func NewChan(channame string) error {
	tx := BC.NewTransaction()
}
func IsExsistChan(name string) error {
	flag := true
	localBlockChain.Traverse(func(block *BC.Block, err error) bool {
		for _, tx := range block.TxInfos {

			//
			if tx.Key == name {
				if tx.DataType == BC.NEW_CHAN {
					flag = false
					return false
				} else if tx.DataType == BC.DEL_CHAN {
					return false
				}
			}

		}
		return true
	})
	if !flag {
		// 找到
		return nil
	}
	return errors.New("不存在的 sharechan: " + name)
}
