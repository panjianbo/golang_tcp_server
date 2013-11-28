package business

import (
	"bytes"
	"errors"
)

/*
            0xAA0x0A == 0xAA
            0xAA0x0B == 0xBB
            0xAA0x0E == 0xEE
*/
func FrameUnescape(input []byte)([]byte, error){
	length := len(input) 
	output := make([]byte , length)
	index := 0
	for i:=0;i<length; i++ {
		if input[i] == 0xaa {
			i++;
			if i >= length {
				return nil, errors.New("unescape error") 
			}
			if input[i] == 0x0a{
				output[index] = 0xaa
			}else if input[i] == 0x0b{
				output[index] = 0xbb
			}else if input[i] == 0x0e{
				output[index] = 0xee
			}else{
				return nil, errors.New("unescape error")
			}
		}else{
			output[index] = input[i]
		}
		index ++
	}
	return output[index:], nil
}

func FrameParse(buffer []byte) ([]byte, []byte){
	iBegin := bytes.IndexByte(buffer, 0xbb)
	if iBegin < 0 {
		return nil, nil
	}
	
	iEnd := bytes.IndexByte(buffer, 0xee)
	if iEnd < 0{
		return nil, buffer[iBegin:]
	}else if(iEnd == len(buffer)-1){
		return buffer[iBegin:iEnd], nil
	}
	return buffer[iBegin:iEnd], buffer[iEnd:]
}

func DisposeData(buff []byte)([]byte){
	return nil
}


func AnalysisInputData(buff *bytes.Buffer) ([][]byte, error){
	buffer := buff.Bytes()
	output := make([][]byte, 0)
	for {
		data, left := FrameParse(buffer)
		if data != nil {
			udata, err := FrameUnescape(data)
			if err != nil{
				if resp := DisposeData(udata); resp != nil{
					output = append(output, resp)
				}
			}
		}
		
		if left == nil {
			buff.Reset()
			break
		}
		
		if data != nil {
			buffer = make([]byte, len(left))
			copy(buffer, left)
			continue
		}
		buff.Write(left)
	}
	return output, nil
}