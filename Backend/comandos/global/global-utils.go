package global

func Global_Data(value string) [10]byte {
	temp := make([]byte, 10)
	for i := range temp {
		temp[i] = '\x00'
	}
	copy(temp[:], []byte(value))
	return [10]byte(temp)
}

func Global_ID(value string) [4]byte {
	temp := make([]byte, 4)
	for i := range temp {
		temp[i] = '\x00'
	}
	copy(temp[:], []byte(value))
	return [4]byte(temp)
}
