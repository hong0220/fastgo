package utils

func Diff(dbIds []int, requestIds []int) (createIds []int, updateIds []int, deleteIds []int) {
	// 创建
	createIds = []int{}
	for i := range requestIds {
		if requestIds[i] < 0 {
			createIds = append(createIds, requestIds[i])
		}
	}

	updateIds = []int{}
	for i := range dbIds {
		for j := range requestIds {
			if dbIds[i] == requestIds[j] {
				updateIds = append(updateIds, dbIds[i])
				break
			}
		}
	}

	deleteIds = []int{}
	for i := range dbIds {
		var exist = false
		for j := range requestIds {
			if dbIds[i] == requestIds[j] {
				exist = true
			}
		}
		// 不存在
		if !exist {
			deleteIds = append(deleteIds, dbIds[i])
		}
	}

	return
}
