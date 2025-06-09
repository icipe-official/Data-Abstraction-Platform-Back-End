package metadatamodel

func FilterFieldGroups(mmGroup any, callback func(property map[string]any) bool) any {
	mmGroupMap, err := GetFieldGroupMap(mmGroup)
	if err != nil {
		return mmGroup
	}

	mmGroupFields, err := GetGroupFields(mmGroupMap)
	if err != nil {
		return mmGroup
	}
	mmGroupReadOrderOfFields, err := GetGroupReadOrderOfFields(mmGroup)
	if err != nil {
		return mmGroup
	}

	readOrderOfFieldsToExclude := make([]int, 0)
	for fgKeySuffixIndex, fgKeySuffix := range mmGroupReadOrderOfFields {
		fgKeySuffixString, err := GetValueAsString(fgKeySuffix)
		if err != nil {
			return mmGroup
		}

		fgMap, err := GetFieldGroupMap(mmGroupFields[fgKeySuffixString])
		if err != nil {
			return mmGroup
		}

		if !callback(fgMap) {
			readOrderOfFieldsToExclude = append(readOrderOfFieldsToExclude, fgKeySuffixIndex)
			delete(mmGroupMap[FIELD_GROUP_PROP_GROUP_FIELDS].([]any)[0].(map[string]any), fgKeySuffixString)
		} else {
			if _, err := GetGroupFields(fgMap); err == nil {
				if _, err := GetGroupReadOrderOfFields(fgMap); err == nil {
					mmGroupMap[FIELD_GROUP_PROP_GROUP_FIELDS].([]any)[0].(map[string]any)[fgKeySuffixString] = FilterFieldGroups(fgMap, callback)
				}
			}
		}
	}

	for _, rDeleteIndex := range readOrderOfFieldsToExclude {
		mmGroupMap[FIELD_GROUP_PROP_GROUP_READ_ORDER_OF_FIELDS] = append(mmGroupMap[FIELD_GROUP_PROP_GROUP_READ_ORDER_OF_FIELDS].([]any)[:rDeleteIndex], mmGroupMap[FIELD_GROUP_PROP_GROUP_READ_ORDER_OF_FIELDS].([]any)[rDeleteIndex+1:])
	}

	return mmGroupMap
}

func ForEachFieldGroup(mmGroup any, callback func(property map[string]any) bool) {
	mmGroupMap, err := GetFieldGroupMap(mmGroup)
	if err != nil {
		return
	}

	mmGroupFields, err := GetGroupFields(mmGroupMap)
	if err != nil {
		return
	}
	mmGroupReadOrderOfFields, err := GetGroupReadOrderOfFields(mmGroup)
	if err != nil {
		return
	}

	for _, fgKeySuffix := range mmGroupReadOrderOfFields {
		fgKeySuffixString, err := GetValueAsString(fgKeySuffix)
		if err != nil {
			return
		}

		fgMap, err := GetFieldGroupMap(mmGroupFields[fgKeySuffixString])
		if err != nil {
			return
		}

		if value := callback(fgMap); value {
			return
		}

		if _, err := GetGroupFields(fgMap); err == nil {
			if _, err := GetGroupReadOrderOfFields(fgMap); err == nil {
				ForEachFieldGroup(fgMap, callback)
			}
		}
	}
}

func MapFieldGroups(mmGroup any, callback func(property map[string]any) any) any {
	mmGroupMap, err := GetFieldGroupMap(mmGroup)
	if err != nil {
		return mmGroup
	}

	mmGroupFields, err := GetGroupFields(mmGroupMap)
	if err != nil {
		return mmGroup
	}
	mmGroupReadOrderOfFields, err := GetGroupReadOrderOfFields(mmGroup)
	if err != nil {
		return mmGroup
	}

	for _, fgKeySuffix := range mmGroupReadOrderOfFields {
		fgKeySuffixString, err := GetValueAsString(fgKeySuffix)
		if err != nil {
			return mmGroup
		}

		fgMap, err := GetFieldGroupMap(mmGroupFields[fgKeySuffixString])
		if err != nil {
			return mmGroup
		}

		mmGroupMap[FIELD_GROUP_PROP_GROUP_FIELDS].([]any)[0].(map[string]any)[fgKeySuffixString] = callback(fgMap)
		if _, err := GetGroupFields(mmGroupMap[FIELD_GROUP_PROP_GROUP_FIELDS].([]any)[0].(map[string]any)[fgKeySuffixString]); err == nil {
			if _, err := GetGroupReadOrderOfFields(mmGroupMap[FIELD_GROUP_PROP_GROUP_FIELDS].([]any)[0].(map[string]any)[fgKeySuffixString]); err == nil {
				mmGroupMap[FIELD_GROUP_PROP_GROUP_FIELDS].([]any)[0].(map[string]any)[fgKeySuffixString] = MapFieldGroups(mmGroupMap[FIELD_GROUP_PROP_GROUP_FIELDS].([]any)[0].(map[string]any)[fgKeySuffixString], callback)
			}
		}
	}

	return mmGroupMap
}
