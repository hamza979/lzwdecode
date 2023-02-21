package lzwdecode

import (
	"encoding/binary"
	"fmt"
	"os"
)

func resetDic() map[uint16][]byte {
	m := make(map[uint16][]byte)
	for i := 0; i < 256; i++ {
		m[uint16(i)] = []byte{byte(i)}
	}
	return m
}


func processNextCode(m *map[uint16][]byte, current_num uint16, characters []byte, new_val *uint16) []byte {
	var entry []byte
	//checking whether code is present in dictionary
	val, ok := (*m)[current_num]
	if ok {
		entry = val
	} else {
		//if not present, indicates duplicate character
		entry = append(characters, characters[0:1]...)
	}
	//have to create deep copy for concatenated dictionary entry to prevent modifying previous entries 
	cpy := make([]byte, len(characters)+1)
	copy(cpy, append(characters, entry[0:1]...))
	(*m)[*new_val] = cpy
	characters = entry
	*new_val++
	//check if dict full
	if *new_val == 4096 {
		*new_val = 256
		*m = resetDic()
	}
	return entry

}
func DecodeFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	const codeWidth = 12
	m := resetDic()
	new_val := uint16(256)
	var buf [3]byte
	//read first 3 bytes
	_, err = file.Read(buf[:])
	//extract the two 12 bit codes
	chunk1 := binary.BigEndian.Uint16(buf[:2]) >> 4
	chunk2 := binary.BigEndian.Uint16(buf[1:]) & 4095
	current_num := chunk2
	characters := []byte{byte(chunk1)}
	result := []byte{byte(chunk1)}
	for {
		//take next 3 bytes
		_, err := file.Read(buf[:])
		if err != nil {
			break
		}
		//extract 12 bit codes
		chunk1 := binary.BigEndian.Uint16(buf[:2]) >> 4
		chunk2 := binary.BigEndian.Uint16(buf[1:]) & 4095
		next_num := chunk1
		//passing pointer to map so that it can still be reassigned in resetDic
		characters = processNextCode(&m, current_num, characters, &new_val)
		result = append(result, characters...)
		current_num = next_num
		next_num = chunk2
		characters = processNextCode(&m, current_num, characters, &new_val)
		result = append(result, characters...)
		current_num = next_num
	}
	//assuming input file will be of the format file.xyz.z
	outputFile, outputErr := os.Create(filePath[:len(filePath)-2])
	if outputErr != nil {
		fmt.Println(outputErr)
		return
	}
	defer outputFile.Close()
	_, err = outputFile.Write(result)
	if outputErr != nil {
		panic(outputErr)
	}
}
