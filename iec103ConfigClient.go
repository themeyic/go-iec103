package iec103

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	//start character
	StartCharacter10 = "10"
	//end character
	Endcharacter = "16"

	StartCharacter68 = "68"
)

type Iec103ConfigClient struct {
	LinkAddress string
	/*
	   FCB(Frame Count Degree):FCB = 0 or 1
	      ——Each time the master sends a new round of "send/confirm" or "request/response"
	   transmission service to the slave station,the FCB is inverted.The master station saves
	   a copy of FCB for each slave station,if no response is received over time,the master station
	   will retransmit,the FCB of the retransmitted message remains unchanged,and the number of
	   retransmissions does not exceed 3 times at most.If the expected response is not received after
	   3 retransmissions, the current round of transmissions service will be terminated
	*/
	FCB int
	/*
	   FCV(Frame count valid bit)
	   FCB = 0 indicates that the change of FCB is invalid, and FCB = 1 indicates that the change of FCB is invalid
	   Send/no answer service,broadcast message does not consider message loss and repeated transmissions no need
	   to change the FCB state, these frames FCV always 0
	*/
	FCV int
	/*
	   ACD(Access bit required)
	   ACD = 1,notify the master station that the slave station has Class 1 data request transmission.
	*/
	//ACD  int
	/*
	   DFC(Data Flow Control Bit) DFC = 0 means the slave station can accept data,DFC = 1 means the
	   slave statio buffer is full and cannot accept new data
	*/
	//DFC  int

	//Type identification
	TYP string

	//Reason for transmission
	COT string

	//Function type
	FUN string

	//Message number
	INF string

	//General classification identification number
	GIN string
}

func (iec103 *Iec103ConfigClient) Initialize(iec Client) string {
	/*
	 This message is used for initialized(sent by the master station)-Reset the communication unit
	      10  		       40             01           41             16
	 Start character	control domain  Link Address  Frame checksum  End character
	*/
	resetTheCommunicationUnit := StartCharacter10 + " 40 " + iec103.LinkAddress + " 41 " + Endcharacter
	slaveBackMessage, err := iec.SendRawFrame(resetTheCommunicationUnit)
	judgmentBitsString := getControlArea(slaveBackMessage, err)
	if judgmentBitsString[2:3] == "1" && judgmentBitsString[3:4] == "0" {
		controlDomain := "01" + strconv.Itoa(iec103.FCB) + strconv.Itoa(iec103.FCV) + "1010"
		if iec103.FCB == 0 {
			iec103.FCB = 1
		} else {
			iec103.FCB = 0
		}
		controlDomain = ConvertBinaryTo16Base(controlDomain)
		slaveBackMessageSecond, _ := iec.SendRawFrame(StartCharacter10 + controlDomain + iec103.LinkAddress + CheckCode(controlDomain+iec103.LinkAddress) + Endcharacter)
		if slaveBackMessageSecond != "" {
			fmt.Println(slaveBackMessageSecond)
		}
	}

	return "Loading Finished !"
}

func getControlArea(judgeString string, err error) string {
	var frameArray []string
	if err == nil {
		frameArray = strings.Split(judgeString, " ")
	}
	test1, _ := strconv.Atoi(frameArray[1])
	you, _ := strconv.ParseUint(strconv.Itoa(test1), 16, 32)
	judgmentBits, err := DecConvertToX(int(you), 2)
	if err != nil {
		fmt.Println(err)
	}
	judgmentBitsString := judgmentBits
	for {
		if len(judgmentBitsString) == 8 {
			break
		}
		judgmentBitsString = "0" + judgmentBitsString
	}
	return judgmentBitsString
}

func (iec103 *Iec103ConfigClient) SummonSecondaryData(iec Client) string {
	summonSecondaryControlDomain := "01" + strconv.Itoa(iec103.FCB) + strconv.Itoa(iec103.FCV) + "1010"
	if iec103.FCB == 0 {
		iec103.FCB = 1
	} else {
		iec103.FCB = 0
	}
	summonSecondaryControlDomain = ConvertBinaryTo16Base(summonSecondaryControlDomain)
	summonSecondaryDataBackFrame, err := iec.SendRawFrame(StartCharacter10 + summonSecondaryControlDomain + iec103.LinkAddress + CheckCode(summonSecondaryControlDomain+iec103.LinkAddress) + Endcharacter)
	judgmentBitsString := getControlArea(summonSecondaryDataBackFrame, err)
	if judgmentBitsString[2:3] == "1" && judgmentBitsString[3:4] == "0" {
		controlDomain := "01" + strconv.Itoa(iec103.FCB+1) + strconv.Itoa(iec103.FCV+1) + "1010"
		controlDomain = ConvertBinaryTo16Base(controlDomain)
		slaveBackMessageSecond, _ := iec.SendRawFrame(StartCharacter10 + controlDomain + iec103.LinkAddress + CheckCode(controlDomain+iec103.LinkAddress) + Endcharacter)
		if iec103.FCB == 0 {
			iec103.FCB = 1
		} else {
			iec103.FCB = 0
		}
		if slaveBackMessageSecond != "" {
			fmt.Println(slaveBackMessageSecond)
		}
	} else if judgmentBitsString[2:3] == "0" {
		fmt.Println("无所召唤的数据")
	} else if judgmentBitsString[3:4] == "1" {
		fmt.Println("子站数据已满不能再接受数据了!")
	}
	return ""
}

func (iec103 *Iec103ConfigClient) MasterStationReadsAnalogQuantity(iec Client, groupNum []int) []float32 {
	var backData []float32
	masterStationReadsAnalogQuantitControlDomain := "01" + strconv.Itoa(iec103.FCB) + "10011"
	if iec103.FCB == 0 {
		iec103.FCB = 1
	} else {
		iec103.FCB = 0
	}
	masterStationReadsAnalogQuantitControlDomain = ConvertBinaryTo16Base(masterStationReadsAnalogQuantitControlDomain)
	resetTheCommunicationUnit := StartCharacter68 + " 0b 0b " + StartCharacter68 + " " + masterStationReadsAnalogQuantitControlDomain + " " + iec103.LinkAddress + " " + iec103.TYP + " 81 " + iec103.COT + " 01 " + iec103.FUN + " " + iec103.INF + " 00 01 09 0e 16"
	slaveBackMessage, err := iec.SendRawFrame(resetTheCommunicationUnit)
	judgmentBitsString := getControlArea(slaveBackMessage, err)
	var markCount int
	if judgmentBitsString[2:3] == "1" && judgmentBitsString[3:4] == "0" {
		controlDomain := "01" + strconv.Itoa(iec103.FCB) + strconv.Itoa(iec103.FCV) + "1010"
		if iec103.FCB == 0 {
			iec103.FCB = 1
		} else {
			iec103.FCB = 0
		}
		controlDomain = ConvertBinaryTo16Base(controlDomain)
		slaveBackMessageSecond, _ := iec.SendRawFrame(StartCharacter10 + controlDomain + iec103.LinkAddress + CheckCode(controlDomain+iec103.LinkAddress) + Endcharacter)
		if slaveBackMessageSecond != "" {
			messageArray := strings.Split(slaveBackMessageSecond, " ")
			messageArray = messageArray[14 : len(messageArray)-2]
			for _, eachGroupNum := range groupNum {
				if markCount = len(messageArray) / 10; eachGroupNum <= len(messageArray)/10 {
					getValue := splitArray(messageArray)[eachGroupNum-1]
					theValue := "0x" + getValue[len(getValue)-1] + getValue[len(getValue)-2] + getValue[len(getValue)-3] + getValue[len(getValue)-4]
					realValue, _ := strconv.ParseUint(theValue, 0, 32)
					finshedvalue := math.Float32frombits(uint32(realValue))
					backData = append(backData, finshedvalue)
				}
			}

			judgeString := messageArray[4]
			judgeString = getControlArea(slaveBackMessage, err)
			for {
				if judgeString[2:3] == "1" && judgeString[3:4] == "0" {
					controlDomainFor := "01" + strconv.Itoa(iec103.FCB) + strconv.Itoa(iec103.FCV) + "1010"
					if iec103.FCB == 0 {
						iec103.FCB = 1
					} else {
						iec103.FCB = 0
					}
					controlDomainFor = ConvertBinaryTo16Base(controlDomainFor)
					slaveBackMessageSecondFor, _ := iec.SendRawFrame(StartCharacter10 + controlDomainFor + iec103.LinkAddress + CheckCode(controlDomainFor+iec103.LinkAddress) + Endcharacter)
					if slaveBackMessageSecondFor != "" {
						if len(strings.Split(slaveBackMessageSecondFor, " ")) == 5 && getControlArea(slaveBackMessageSecondFor, err)[2:3] == "0" && getControlArea(slaveBackMessageSecondFor, err)[3:4] == "0" {
							break
						} else {
							slaveBackMessageSecondFor := strings.Split(slaveBackMessageSecondFor, " ")
							slaveBackMessageSecondFor = slaveBackMessageSecondFor[14 : len(slaveBackMessageSecondFor)-2]

							for _, eachGroupNum := range groupNum {
								if eachGroupNum > markCount && eachGroupNum <= markCount+len(slaveBackMessageSecondFor)/10 {
									getValue := splitArray(slaveBackMessageSecondFor)[eachGroupNum-markCount-1]
									theValue := "0x" + getValue[len(getValue)-1] + getValue[len(getValue)-2] + getValue[len(getValue)-3] + getValue[len(getValue)-4]
									realValue, _ := strconv.ParseUint(theValue, 0, 32)
									finshedvalue := math.Float32frombits(uint32(realValue))
									backData = append(backData, finshedvalue)
								}
							}
						}
					}

				} else if judgeString[2:3] == "0" {
					fmt.Println("无所召唤的数据")
					break
				} else if judgeString[3:4] == "1" {
					fmt.Println("子站数据已满不能再接受数据了!")
					break
				}
			}

		}
	} else if judgmentBitsString[2:3] == "0" {
		fmt.Println("无所召唤的数据")
	} else if judgmentBitsString[3:4] == "1" {
		fmt.Println("子站数据已满不能再接受数据了!")
	}
	return backData
}

func DecConvertToX(n, num int) (string, error) {
	if n < 0 {
		return strconv.Itoa(n), errors.New("只支持正整数")
	}
	if num != 2 && num != 8 && num != 16 {
		return strconv.Itoa(n), errors.New("只支持二、八、十六进制的转换")
	}
	result := ""
	h := map[int]string{
		0:  "0",
		1:  "1",
		2:  "2",
		3:  "3",
		4:  "4",
		5:  "5",
		6:  "6",
		7:  "7",
		8:  "8",
		9:  "9",
		10: "A",
		11: "B",
		12: "C",
		13: "D",
		14: "E",
		15: "F",
	}
	for ; n > 0; n /= num {
		lsb := h[n%num]
		result = lsb + result
	}
	return result, nil
}

func splitArray(arr []string) [][]string {
	//circle time
	finshedArray := make([][]string, 0)
	for i := 0; i < len(arr); i += 10 {
		finshedArray = append(finshedArray, arr[i:i+10])

	}
	return finshedArray
}

func StringToIntArray(input string) []int {
	output := []int{}
	for _, v := range input {
		output = append(output, int(v)-48)
	}
	for i, j := 0, len(output)-1; i < j; i, j = i+1, j-1 {
		output[i], output[j] = output[j], output[i]
	}
	return output
}

func ConvertBinaryTo16Base(controlDomain string) string {
	sixthBaseMap := map[int]string{
		0:  "0",
		1:  "1",
		2:  "2",
		3:  "3",
		4:  "4",
		5:  "5",
		6:  "6",
		7:  "7",
		8:  "8",
		9:  "9",
		10: "a",
		11: "b",
		12: "c",
		13: "d",
		14: "e",
		15: "f",
	}
	controlDomainArray := StringToIntArray(controlDomain)
	endValue1 := 0
	endValue2 := 0
	for i := 0; i < len(controlDomainArray)/2; i++ {
		endValue1 += controlDomainArray[i] * int(math.Pow(2, float64(i)))
	}
	mark := 0
	for j := len(controlDomainArray) / 2; j < len(controlDomainArray); j++ {
		endValue2 += controlDomainArray[j] * int(math.Pow(2, float64(mark)))
		mark++
	}
	leftValue := sixthBaseMap[endValue2]
	rightValue := sixthBaseMap[endValue1]

	return leftValue + rightValue
}

//计算出校验码
func CheckCode(data string) string {
	data = strings.ReplaceAll(data, " ", "")
	total := 0
	length := len(data)
	num := 0
	for num < length {
		s := data[num : num+2]
		//16进制转换成10进制
		totalMid, _ := strconv.ParseUint(s, 16, 32)
		total += int(totalMid)
		num = num + 2
	}
	//将校验码前面的所有数通过16进制加起来转换成10进制，然后除256区余数，然后余数转换成16进制，得到的就是校验码
	mod := total % 256
	hex, _ := DecConvertToX(mod, 16)
	len := len(hex)
	//如果校验位长度不够，就补0，因为校验位必须是要2位
	if len < 2 {
		hex = "0" + hex
	}
	return strings.ToLower(hex)
}
