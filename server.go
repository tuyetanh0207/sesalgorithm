package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"strconv"
	"math/rand"
	"time"
	"log"
)

type Config struct {
	Ports []int `json:"ports"`
}

type Timestamp struct {
	Time int
}

type Object struct {
	Timestamp Timestamp
	ProcessID int
}

type ReceivedItem struct {
	ProcessID int
	Timestamps []int
	Content string
}
type BufferItem struct {
	ReceivedList []ReceivedItem
	Timestamp []int
	ProcessID int
}

type Message struct {
	Timestamps    []int
	ReceivedArray []ReceivedItem
	Sender        int
	Content string
}

func loadConfig(filename string) (Config, error) {
	var config Config

	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		if err := decoder.Decode(&config); err == io.EOF {
			break
		} else if err != nil {
			return config, err
		}
	}

	return config, nil
}

func sendMessage(senderID, receiverID int, timestamps []int, receivedArray []ReceivedItem, chs []chan Message, messageName string) {
	//fmt.Printf("In sendMessage function: prepare to implement sendMessage from process %d to process %d, message is: %s, timestamp is %v\n", senderID, receiverID, messageName, timestamps)
	message := Message{
		Timestamps:    timestamps,
		ReceivedArray: receivedArray,
		Sender:        senderID,
		Content: messageName,
	}
	//fmt.Println("In sendMessage function: create message successfully, message is:", message)
	chs[receiverID] <- message
	//fmt.Println("In sendMessage function: sent to channel of receiver successfully, message is:", message)
}
// if timestamp1 >= timestamp2 returns true, else false
func compareTwoTimestamps(timestamp1 []int, timestamp2 []int) bool {
	//fmt.Println("In compareTwoTimestamps function: timestamp1 is:", timestamp1, " timestamp2 is:", timestamp2)
	tempTimestamp := make([]int, len(timestamp1))
	copy(tempTimestamp, timestamp1)
	for i := 0; i < len(timestamp1); i++ {
            if tempTimestamp[i] < timestamp2[i] {
				//fmt.Println("In compareTwoTimestamps function: timestamp1 is:", timestamp1, " timestamp2 is:", timestamp2, " result is timestamp1 >= timestamp2, return false")

                return false
            }
    }
	//fmt.Println("In compareTwoTimestamps function: timestamp1 is:", timestamp1, " timestamp2 is:", timestamp2, " result is timestamp1 < timestamp2, return true")

	return true
}
func mergeTwoTimestamps(timestamp1 []int, timestamp2 []int, receiverID int) []int {
	//fmt.Println("In mergeTwoTimestamps function: timestamp1 is:", timestamp1, " timestamp2 is:", timestamp2, " receiverID is ", receiverID)
	tempTimestamp1 := make([]int, len(timestamp1))
	copy(tempTimestamp1, timestamp1)
	tempTimestamp1[receiverID]++
	////fmt.Println("In mergeTwoTimestamps function: timestamp1 after incremented is:", timestamp1, " timestamp2 is:", timestamp2, " receiverID is ", receiverID)
	for i := 0; i < len(timestamp1); i++ {
        if timestamp2[i] > tempTimestamp1[i] {
			tempTimestamp1[i] = timestamp2[i]
		}
	}
	//fmt.Println("In mergeTwoTimestamps function: timestamp1 is:", timestamp1, " timestamp2 is:", timestamp2, " receiverID is ", receiverID, " final timestamp after merge is:", tempTimestamp1)

	return tempTimestamp1
}
func checkSentBefore(carriers []ReceivedItem, receiverID int) bool {
	//fmt.Println("In checkSentBefore function: carriers is:", carriers, " receiverID is:", receiverID)

	for i := 0; i < len(carriers); i++ {
        if carriers[i].ProcessID == receiverID {
			//fmt.Println("In checkSentBefore function: carriers is:", carriers, " receiverID is:", receiverID, " result is carriers[i].ProcessID == receiverID, return true")	
            return true
        }
    }
	//fmt.Println("In checkSentBefore function: carriers is:", carriers, " receiverID is:", receiverID, " result is carriers[i].ProcessID!= receiverID, return false")	
    return false

}
func sentBeforeAt(carriers []ReceivedItem, receiverID int) int {
	//fmt.Println("In sentBeforeAt function: carriers is:", carriers, " receiverID is:", receiverID)

    for i := 0; i < len(carriers); i++ {
        if carriers[i].ProcessID == receiverID {
            //fmt.Println("In sentBeforeAt function: carriers is:", carriers, " receiverID is:", receiverID, " result is carriers[i].ProcessID == receiverID, return i")    
            return i
        }
    }
    //fmt.Println("In sentBeforeAt function: carriers is:", carriers, " receiverID is:", receiverID, " result is carriers[i].ProcessID!= receiverID, return -1")    
    return -1
}

func mergeTwoCarriers(senderCarriers []ReceivedItem, receiverCarriers []ReceivedItem) []ReceivedItem {
	tempCarriers := make([]ReceivedItem, len(receiverCarriers))
	copy(tempCarriers, receiverCarriers)
	
	//fmt.Println("In mergeTwoCarriers function: senderCarriers is:", senderCarriers, " receiverCarriers is:", receiverCarriers)
	 for i := 0; i < len(senderCarriers); i++ {
		flag:=false

        for j := 0; j < len(receiverCarriers); j++ {
            if senderCarriers[i].ProcessID == receiverCarriers[j].ProcessID {
				flag=true
				if compareTwoTimestamps(senderCarriers[i].Timestamps, receiverCarriers[j].Timestamps){
					tempCarriers[j].Timestamps = senderCarriers[i].Timestamps
				}
				
            }
        }
		if !flag{
			//fmt.Println("In mergeTwoCarriers function: senderCarriers is:", senderCarriers, " prepare to append receiverCarriers is:", receiverCarriers)
			tempCarriers = append(tempCarriers, senderCarriers[i])
		}
    }
	//fmt.Println("In mergeTwoCarriers function: senderCarriers is:", senderCarriers, " receiverCarriers is:", receiverCarriers, "final senderCarriers is:", senderCarriers)
	return tempCarriers
}
func getTimestampOfProcessInCarrier(processID int, carriers []ReceivedItem) []int{
	//fmt.Println("In getTimestampOfProcessInCarrier function: processID is", processID, "carriers is: ", carriers)

	for i := 0; i < len(carriers); i++ {
        if carriers[i].ProcessID == processID {
			//fmt.Println("In getTimestampOfProcessInCarrier function: processID is", processID, "carriers is: ", carriers, " result is carriers[i].ProcessID == processID, return carriers[i].Timestamps", carriers[i].Timestamps)
            return carriers[i].Timestamps
        }
    }
	//fmt.Println("In getTimestampOfProcessInCarrier function: processID is", processID, "carriers is: ", carriers, " result is carriers[i].ProcessID!= processID, return nil")
    return nil

}
func clearCarriers(carriers []ReceivedItem, processID int, timestamps []int) []ReceivedItem {
	tempCarriers := make([]ReceivedItem, len(carriers))
	copy(tempCarriers, carriers)
    for i := 0; i < len(tempCarriers); i++ {
        if tempCarriers[i].ProcessID == processID {
			if !compareTwoTimestamps(tempCarriers[i].Timestamps, timestamps) {
				tempCarriers=append(tempCarriers[:i],tempCarriers[i+1:]... )
				break
        	}
		}
    }
	for i := 0; i < len(tempCarriers); i++ {
        if len(tempCarriers[i].Timestamps) == 0 {
			tempCarriers=append(tempCarriers[:i],tempCarriers[i+1:]... )
			break
		}
    }
	//fmt.Println("In clearCarriers function: carriers is:", carriers, " processID is:", processID, " timestamps is:", timestamps, "final carriers is:", tempCarriers)
    return tempCarriers
	
}

func minOfBuffer(buffers []BufferItem, processID int, requiredTimestamp []int) (BufferItem, ReceivedItem, int, int) {
	//fmt.Println("In minOfBuffer function: buffers is:", buffers, " processID is:", processID)
	count:=0
	if len(buffers) >= 1 {
		minBuffer := buffers[0]
		minReceivedItem := buffers[0].ReceivedList[0]
		minTimestamp := buffers[0].ReceivedList[0].Timestamps
		outIdx:=-1
		inIdx:=-1
		for i:=0; i<len(buffers); i++ {
			for j:=0; j<len(buffers[i].ReceivedList); j++ {
				if buffers[i].ReceivedList[j].ProcessID==processID && compareTwoTimestamps(buffers[i].ReceivedList[j].Timestamps, requiredTimestamp) {
					if count == 0 {
                    minBuffer = buffers[i]
                    minReceivedItem = buffers[i].ReceivedList[j]
                    minTimestamp = buffers[i].ReceivedList[j].Timestamps
                    outIdx=i
                    inIdx=j
					count++
					continue
					} else if compareTwoTimestamps(buffers[i].ReceivedList[j].Timestamps, minTimestamp){
							minBuffer = buffers[i]
                            minReceivedItem = buffers[i].ReceivedList[j]
                            minTimestamp = buffers[i].ReceivedList[j].Timestamps
                            outIdx=i
                            inIdx=j
					} else {
						continue
					}
					
					
                }
			}
		}
		//fmt.Println("In minOfBuffer function: find out min of buffer, buffer is:", buffers, " processID is:", processID, " minBuffer is: ", minBuffer, " minReceivedItem is:", minReceivedItem, " outIdx is:", outIdx, " inIdx is: ", inIdx)
		return minBuffer, minReceivedItem, outIdx, inIdx
	}
	//fmt.Println("In minOfBuffer function: buffers is:", buffers, " processID is:", processID, "cannot find min buffer, it mean in buffer don't contain processID")

	return BufferItem{}, ReceivedItem{}, -1, -1

}
func removeElement(slice []BufferItem, index int) []BufferItem {
    return append(slice[:index], slice[index+1:]...)
}
// return - 1 when when can find resonable min buffer
func processBuffers(buffers [][]BufferItem, processID int, timestamps [][]int, requiredTimestamp []int, carriers [][]ReceivedItem ) (int, string) {
	//fmt.Println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers)
	tempBuffer := make([]BufferItem,len(buffers[processID]))
	copy(tempBuffer,buffers[processID])
	minBuffer, minReceivedItem, outIdx, inIdx := minOfBuffer(tempBuffer, processID, requiredTimestamp)
	
	if outIdx < 0  || inIdx < 0 {
		return -1, "Invalid min buffer index"
	}
	//fmt.Println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx)
	
	////fmt.Println("In processBuffers function: minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx)
	message := Message{
		Timestamps:    minReceivedItem.Timestamps,
        ReceivedArray: minBuffer.ReceivedList,
        Sender:        minBuffer.ProcessID,
        Content: minReceivedItem.Content,
	}
	//fmt.Println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx, " prepare implement receiveMessageProcess in it")
	buffers[processID]=removeElement(buffers[processID], outIdx)
	res := receiveMessageProcessing(timestamps, buffers, processID, minReceivedItem.ProcessID, carriers, message)
	//fmt.Println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx, " implemented receiveMessageProcess in it, res is ", res)
	
	// delete buffer sent
	
	//fmt.Println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx, " buffer after cleared is", buffers[processID])
	
	if res ==1 {
		//fmt.Println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx, " implemented receiveMessageProcess sucessfully in processBuffer ", res)
		return 1, "Process buffer of process " + strconv.Itoa(processID) + "from process " +strconv.Itoa(minBuffer.ProcessID)+ " .Message name is: " + minReceivedItem.Content +" successfully."
	} else {
		//fmt.Println("In processBuffers function: buffers is:", buffers, " processID is:", processID, "timestamps is:", timestamps, " carrier is:", carriers, "minBuffer is:", minBuffer, " minReceivedItem is:", minReceivedItem, " carrier is:", carriers, " outIdx is:", outIdx, " inIdx is:", inIdx, " implemented receiveMessageProcess but unsucessfully, in processBuffer ", res)
		return 0, "Error in receiving message in process buffer  of process " + strconv.Itoa(processID) + "from process " +strconv.Itoa(minBuffer.ProcessID)+ " .Message name is: " + minReceivedItem.Content +" ."
	}
}
func writeToLogFile(fileNumber int, message string) error {
	// Tạo tên tệp dựa trên số thứ tự
	fileName := fmt.Sprintf("logfile%d.txt", fileNumber)

	// Mở hoặc tạo tệp nhật ký (sử dụng flag os.O_APPEND để ghi tiếp vào tệp đã tồn tại)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Ghi thông điệp vào tệp nhật ký
	_, err = fmt.Println(file, "%s\n", message)
	if err != nil {
		return err
	}

	return nil
}
func receiveMessageProcessing(timestamps [][]int, buffers [][]BufferItem, currReceiverID int, senderID int, carriers [][]ReceivedItem, message Message) int {

	////fmt.Println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message)
	//fmt.Println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to merge 2 timestamps, timestamp1 is", timestamps[currReceiverID], " timestamp2 is", message.Timestamps, " currReceiverID is", currReceiverID)
	//tempTimestamp := make([]int, len(timestamps[currReceiverID]))
	tempTimestamp := timestamps[currReceiverID]
	
	timestamps[currReceiverID] = mergeTwoTimestamps(tempTimestamp, message.Timestamps, currReceiverID)
	//fmt.Println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to merge 2 timestamps, timestamp of receiver is", tempTimestamp, " timestamp2 is", message.Timestamps, " currReceiverID is", currReceiverID, " timestamp after merge is", timestamps[currReceiverID])
	

	//fmt.Println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to merge 2 carriers, carrier 1 is", carriers[currReceiverID], " carrier2 is", message.ReceivedArray)
	
	carriers[currReceiverID]= mergeTwoCarriers(carriers[currReceiverID], message.ReceivedArray)
	//fmt.Println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to merge 2 carriers, carrier 1 is", carriers[currReceiverID], " carrier2 is", message.ReceivedArray, ". carrier after merge is", carriers[currReceiverID])
	
	//fmt.Println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to clear carrier, carrier is", carriers[currReceiverID], " senderID is", senderID, " timestamp is: ", message.Timestamps)
	
	carriers[currReceiverID] = clearCarriers(carriers[currReceiverID], senderID, message.Timestamps)
	//fmt.Println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to clear carrier, carrier is", carriers[currReceiverID], " senderID is", senderID, " timestamp is: ", message.Timestamps, " carrier after clear is: ", carriers[currReceiverID])
	
	//fmt.Println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to process buffer, buffer is", buffers[currReceiverID], "timestamps is", timestamps, "carcarriers is: ", carriers[currReceiverID])
	
	resProcessBuffer, _ :=processBuffers(buffers, currReceiverID, timestamps, tempTimestamp, carriers)
	//fmt.Println("In ReceiveMessageProcessing function:", "timestamps is", timestamps, "buffers is", buffers, "currReceiverID is", currReceiverID, "senderID is",senderID, "carcarriers is: ", carriers ,"message is", message, ". Prepare to process buffer, buffer is", buffers[currReceiverID], "timestamps is", timestamps, "carcarriers is: ", carriers[currReceiverID], " resBufferProcess is", resProcessBuffer, str)

	if resProcessBuffer==1 {
		//fmt.Printf("In ReceiveMessageProcessing function: Process %d received message from Process %d: successfully! Content is: %s., timestamp is %v .\n", currReceiverID, senderID, message.Content, message.Timestamps)
		return 1
	} else {
		return 0
	}
}

type CustomLogger struct {
	consoleLogger *log.Logger
	fileLogger    *log.Logger
}

func NewCustomLogger(logFileName string) (*CustomLogger, error) {
	// Open or create the log file
	file, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Create a console logger that writes to standard output
	consoleLogger := log.New(os.Stdout, "", log.LstdFlags)

	// Create a file logger that writes to the opened log file
	fileLogger := log.New(file, "", log.LstdFlags)

	return &CustomLogger{
		consoleLogger: consoleLogger,
		fileLogger:    fileLogger,
	}, nil
}

func (c *CustomLogger) PrintLog(index int, message string) {
	// Log to console
	c.consoleLogger.Println(message)

	// Log to file based on index
	if index >= 0 && index < len(loggers) {
		loggers[index].fileLogger.Println(message)
	} else {
		// Handle index out of range error or log to a default file
		c.consoleLogger.Println("Invalid log index:", index)
	}
}
var loggers []*CustomLogger

func main() {
	//fmt.Println("Start program.")
	
	
	//fmt.Println("Log files created successfully in 'logfiles' directory.")
	config, err := loadConfig("config.json")
	if err != nil {
		//fmt.Println("Error reading config file:", err)
		return
	}

	fmt.Println("Ports:", config.Ports)
	if _, err := os.Stat("logfiles"); os.IsNotExist(err) {
		// Thư mục không tồn tại, tạo thư mục mới
		err := os.Mkdir("logfiles", os.ModePerm)
		if err != nil {
			//fmt.Printf("Không thể tạo thư mục: %v\n", err)
			return
		}
		//fmt.Println("Thư mục đã được tạo.")
	} else {
		//fmt.Println("Thư mục đã tồn tại.")
	}
	 os.Chdir("logfiles")
	//var loggers []*CustomLogger
	for i := 0; i < len(config.Ports); i++ {
		logFileName := fmt.Sprintf("logfile%d.txt", i)

		// Kiểm tra xem tệp đã tồn tại hay không
		if _, err := os.Stat(logFileName); err == nil {
			// Tệp đã tồn tại, xóa nó
			err := os.Remove(logFileName)
			if err != nil {
				//fmt.Printf("Không thể xóa tệp %s: %v\n", logFileName, err)
				return
			}
			//fmt.Printf("Tệp %s đã được xóa.\n", logFileName)
		}

		// Tạo tệp mới
		customLogger, err := NewCustomLogger(logFileName)
		if err != nil {
			//fmt.Printf("Lỗi khi tạo logger cho tệp %s: %v\n", logFileName, err)
			return
		}

		// Thêm logger vào danh sách
		loggers = append(loggers, customLogger)

	}

	// Write messages to specific log files based on index
	for i := 0; i < len(config.Ports); i++ {
		loggers[i].PrintLog(i, fmt.Sprintf("Log message for port %d", i+8000))
	}


	var wg sync.WaitGroup

	// Initialize channels and timestamps for processes
	channels := make([]chan Message, len(config.Ports))
	buffers := make([][]BufferItem, len(config.Ports))
	carriers := make([][]ReceivedItem, len(config.Ports))
	timestamps := make([][]int, len(config.Ports))
	for i := 0; i < len(config.Ports); i++ {
		channels[i] = make(chan Message)
		buffers[i] = make([]BufferItem, 0)
		carriers[i] = make([]ReceivedItem, 0)
		timestamps[i] = make([]int, len(config.Ports))
		for j := range timestamps[i] {
			timestamps[i][j] = 0
		}

		////fmt.Println("carriers new:", carriers[i])
		////fmt.Println("timestamps: %d ", i, timestamps)
	}

	// Start processes based on ports from config
	for _, port := range config.Ports {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			for i := 0; i < len(config.Ports); i++ {
				if i != port-8000 {
					go func(senderID, receiverID int) {
						for {
							message := <-channels[receiverID]
							// Process received message logic
							tempSenderID:= message.Sender - 8000
							//fmt.Printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. prepare process.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray)
							loggers[receiverID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. prepare process.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray))
							// check if carrier of sending processed has been sent before
							//loggers[receiverID].PrintLog(tempSenderID, fmt.Sprintln())
							if (!checkSentBefore(message.ReceivedArray, receiverID)) {
								//fmt.Printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray)
								loggers[receiverID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray))
								ress:= receiveMessageProcessing(timestamps, buffers, receiverID, tempSenderID,carriers, message)
								if (ress==1){
									loggers[receiverID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept., receive message successfully . End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray))
									//fmt.Printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept., receive message successfully . End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray) 
								} else {
									loggers[receiverID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept., but receive message unsuccessfully.  End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray))
									//fmt.Printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentbefore, can accept., but receive message unsuccessfully.  End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray) 
								}

							} else{
							// compare 2 timestamps
							////fmt.Println("checksentbefore: ",checkSentBefore(message.ReceivedArray, receiverID) )
							timestampsInCarrier := getTimestampOfProcessInCarrier(receiverID, message.ReceivedArray)
							//fmt.Printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentBefore, prepare to process timestamp, timestampIncarrier is: %v.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier ) //
							loggers[receiverID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after checksentBefore, prepare to process timestamp, timestampIncarrier is: %v.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier))
							if (timestampsInCarrier != nil) {
								
								if compareTwoTimestamps(timestamps[receiverID], timestampsInCarrier) {
									
									res:= receiveMessageProcessing(timestamps, buffers, receiverID, tempSenderID,carriers, message)
									if (res==1){
										loggers[receiverID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is acceptable, receive message successfully %v. End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier))
										//fmt.Printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is acceptable, receive message successfully %v. End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier) 
									} else {
										loggers[receiverID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is acceptable, receive message successfully %v. End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier))
										//loggers[receiverID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is acceptable, but receive message unsuccessfully %v.  End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier ))
										//fmt.Printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is acceptable, but receive message unsuccessfully %v.  End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier ) 
									}
							
								} else { // add received item to buffer
									// buffer if timestamp is not correct
									
									buffers[receiverID] = append(buffers[receiverID], BufferItem{Timestamp: message.Timestamps, ReceivedList:message.ReceivedArray})
									////fmt.Println("Buffer:", buffers[receiverID])
									loggers[receiverID].PrintLog(tempSenderID, fmt.Sprintf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is %v unacceptable, buffer message, current buffer is: %v. End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier, buffers[receiverID] ))
									//fmt.Printf("In receiving process: Process %d received message from Process %d:, message is: %s, timestamp is %v, receivedArray is: %v. after compare and timestamp is %v unacceptable, buffer message, current buffer is: %v. End receiving.\n", receiverID, message.Sender, message.Content, message.Timestamps, message.ReceivedArray, timestampsInCarrier, buffers[receiverID] ) 
								}
							}
							
							}
							
						}
					
					}(i, port-8000)
				}
			}
			fmt.Println("Complete receiving section.")

			
			for i := 0; i < len(config.Ports); i++ {
				for j := 0; j < 150; j++ {
					receiverID := (port + j -1) % len(config.Ports)
					senderID := port - 8000
					
					
					if (receiverID != senderID) {
						////fmt.Println("In sending section:")
						// Chỉ cập nhật timestamps tại vị trí sender tương ứng
						tempTimestamp:=timestamps[senderID]
						//tempTimestamp[senderID]= tempTimestamp[senderID] + 1
						newTimestamp := make([]int, len(tempTimestamp))
						newCarrier := make([]ReceivedItem, len(carriers[senderID]))
						copy(newTimestamp, tempTimestamp)
						copy(newCarrier, carriers[senderID])
						newTimestamp[senderID] = newTimestamp[senderID] + 1
						randGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))			
						messageName := "Message " + strconv.Itoa(senderID*len(config.Ports)+receiverID + i + j + randGenerator.Intn(1000) + randGenerator.Intn(1000))
						//fmt.Printf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v \n", senderID, receiverID, newTimestamp, messageName, carriers[senderID])
						loggers[senderID].PrintLog(senderID, fmt.Sprintf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v \n", senderID, receiverID, newTimestamp, messageName, carriers[senderID]))
						timestamps[senderID][senderID]=newTimestamp[senderID]
						sendMessage(port, receiverID, newTimestamp, newCarrier, channels, messageName)
						
						newItem := ReceivedItem{
							ProcessID:  receiverID,
							Timestamps: newTimestamp,
							Content:    messageName,
						}
						idx := sentBeforeAt(carriers[senderID], receiverID) 
						if idx >=0 && idx<len(carriers[senderID]) {
							////fmt.Println("idx of carrier in if block: ", idx)
							newCarrierItem := ReceivedItem{
								ProcessID:  receiverID,
								Timestamps: newTimestamp,
								Content: carriers[senderID][idx].Content,
							}
							if(len(newCarrierItem.Timestamps)<idx && idx>=0){
			
								carriers[senderID][idx] = newCarrierItem
								//fmt.Printf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v, just replace carrier %v at index %d, current carrier is: %v \n", senderID, receiverID, timestamps[senderID], messageName, carriers[senderID], newCarrierItem, idx, carriers[senderID])
								loggers[senderID].PrintLog(senderID, fmt.Sprintf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v, just replace carrier %v at index %d, current carrier is: %v \n", senderID, receiverID, timestamps[senderID], messageName, carriers[senderID], newCarrierItem, idx, carriers[senderID]))
								//loggers[i].PrintLog(i, fmt.Sprintln())
							}
							
						
						} else {
							if(len(newItem.Timestamps)>=0){
								carriers[senderID] = append(carriers[senderID], newItem)
								//fmt.Printf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v, just append carrier %v at index %d, current carrier is: %v  \n", senderID, receiverID, timestamps[senderID], messageName, carriers[senderID], newItem, idx, carriers[senderID])
								//
								loggers[senderID].PrintLog(senderID, fmt.Sprintf("In sending section: From process %d send to process %d, timestamp is: %v, messageName is: %s, carrier is: %v, just append carrier %v at index %d, current carrier is: %v  \n", senderID, receiverID, timestamps[senderID], messageName, carriers[senderID], newItem, idx, carriers[senderID]))
							
							}
						}
					
					}
					
				}
			}
			//fmt.Println("Complete sending section.")
		}(port)
	}

	wg.Wait()
}
