package Type

import (
	"bytes"
	"go_code/基于区块链的非关系型数据库/util"
)

func ConvertToSTRING(v []byte) string {
	return string(v)
}
func ConvertToINT32(v []byte) int32 {
	return util.BytesToInt32(v)
}
func ConvertToINT64(v []byte) int64 {
	return util.BytesToInt64(v)
}
func ConvertToSTRING_ARRAY(v []byte) []string {
	barr := bytes.Split(v, []byte{byte(0)})
	vstring := []string{}
	for _, v := range barr {
		vstring = append(vstring, string(v))
	}
	return vstring
}

func ConvertToINT32_ARRAY(v []byte) []int32 {
	l := len(v)
	arr := []int32{}
	for i := 0; i < l; i += 4 {
		arr = append(arr, util.BytesToInt32(v[i:i+4]))
	}
	return arr
}
func ConvertToINT64_ARRAY(v []byte) []int64 {
	l := len(v)
	arr := []int64{}
	for i := 0; i < l; i += 8 {
		arr = append(arr, util.BytesToInt64(v[i:i+8]))
	}
	return arr
}
func ConvertToSTRING_SET(v []byte) []string {
	strarr := ConvertToSTRING_ARRAY(v)
	for i := 1; i < len(strarr); i++ {
		tempi := i
		tempv := strarr[i]
		for ; strarr[tempi] < strarr[tempi-1] && tempi > 0; tempi-- {
			strarr[tempi] = strarr[tempi-1]
		}
		strarr[tempi] = tempv
	}
	restr := []string{}
	for i := 0; i < len(strarr); i++ {
		if i == len(strarr)-1 {
			restr = append(restr, strarr[i])
			break
		}
		if strarr[i] == strarr[i+1] {
			continue
		}
		restr = append(restr, strarr[i])
	}
	return restr
}
func ConvertToINT32_SET(v []byte) []int32 {
	strarr := ConvertToINT32_ARRAY(v)
	for i := 1; i < len(strarr); i++ {
		tempi := i
		tempv := strarr[i]
		for ; strarr[tempi] < strarr[tempi-1] && tempi > 0; tempi-- {
			strarr[tempi] = strarr[tempi-1]
		}
		strarr[tempi] = tempv
	}
	restr := []int32{}
	for i := 0; i < len(strarr); i++ {
		if i == len(strarr)-1 {
			restr = append(restr, strarr[i])
			break
		}
		if strarr[i] == strarr[i+1] {
			continue
		}
		restr = append(restr, strarr[i])
	}
	return restr
}
func ConvertToINT64_SET(v []byte) []int64 {
	strarr := ConvertToINT64_ARRAY(v)
	for i := 1; i < len(strarr); i++ {
		tempi := i
		tempv := strarr[i]
		for ; strarr[tempi] < strarr[tempi-1] && tempi > 0; tempi-- {
			strarr[tempi] = strarr[tempi-1]
		}
		strarr[tempi] = tempv
	}
	restr := []int64{}
	for i := 0; i < len(strarr); i++ {
		if i == len(strarr)-1 {
			restr = append(restr, strarr[i])
			break
		}
		if strarr[i] == strarr[i+1] {
			continue
		}
		restr = append(restr, strarr[i])
	}
	return restr
}
