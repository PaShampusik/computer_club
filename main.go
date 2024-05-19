package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Структура события
type Event struct {
	Time   time.Time
	ID     int
	Client string
	Table  int
	Error  string
}

// Структура клуба
type Club struct {
	Tables       int
	Open         time.Time
	Close        time.Time
	Price        int
	Events       []Event
	Clients      map[string]int
	TablesStatus map[int]string
	WaitingQueue []string
}

// Функция для проверки наличия свободных столов
func (c *Club) isTableFreeExists() bool {
	for _, status := range c.TablesStatus {
		if status == "" {
			return true
		}
	}
	return false
}

// Функция для создания события с ошибкой
func createErrorEvent(message string, time time.Time) Event {
	event := Event{
		Time:  time,
		ID:    13,
		Error: message,
	}
	return event
}

// Функция парсинга события
func (e *Event) parseEvent(line string) error {
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		return fmt.Errorf("invalid event format: %s", line)
	}
	t, err := time.Parse("15:04", parts[0])
	if err != nil {
		return fmt.Errorf("invalid time format: %s", parts[0])
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid event ID: %s", parts[1])
	}
	e.Time = t
	e.ID = id
	if id == 13 {
		e.Error = parts[2]
	} else {
		e.Client = parts[2]
		if id == 2 || id == 12 {
			table, err := strconv.Atoi(parts[3])
			if err != nil {
				return fmt.Errorf("invalid table number: %s", parts[3])
			}
			e.Table = table
		}
	}
	return nil
}

// Функция для парсинга входного текстового файла
func (c *Club) parseClub(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("can't open file: %s", filename)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	// Парсинг кол-ва столов
	scanner.Scan()
	lineNumber++
	tables, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return fmt.Errorf("invalid tables number on line %d: %s", lineNumber, scanner.Text())
	}
	c.Tables = tables
	c.TablesStatus = make(map[int]string, tables)
	for i := 1; i <= tables; i++ {
		c.TablesStatus[i] = ""
	}

	// Парсинг времени открытия и закрытия клуба
	scanner.Scan()
	lineNumber++
	parts := strings.Split(scanner.Text(), " ")
	open, err := time.Parse("15:04", parts[0])
	if err != nil {
		return fmt.Errorf("invalid open time on line %d: %s", lineNumber, parts[0])
	}
	close, err := time.Parse("15:04", parts[1])
	if err != nil {
		return fmt.Errorf("invalid close time on line %d: %s", lineNumber, parts[1])
	}
	c.Open = open
	c.Close = close

	// Парсинг цены за час
	scanner.Scan()
	lineNumber++
	price, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return fmt.Errorf("invalid price on line %d: %s", lineNumber, scanner.Text())
	}
	c.Price = price

	c.Clients = make(map[string]int)

	// Парсинг событий
	c.Events = make([]Event, 0)
	for scanner.Scan() {
		lineNumber++
		var event Event
		if err := event.parseEvent(scanner.Text()); err != nil {
			return fmt.Errorf("can't parse event on line %d: %s", lineNumber, scanner.Text())
		}
		c.Events = append(c.Events, event)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("can't read file: %s", filename)
	}

	return nil
}

// Функция для обработки каждого события
func (c *Club) ProcessEvents() {

	// Выводим время открытия компьютерного клуба
	fmt.Println(c.Open.Format("15:04"))

	// Новым массив событий, для сбора записи ивентов и ошибок
	var newEvents = []Event{}

	// Итерация по всем событиям
	for _, event := range c.Events {
		newEvents = append(newEvents, event)
		if event.ID == 1 {
			if _, ok := c.Clients[event.Client]; ok {
				event := createErrorEvent("YouShallNotPass", event.Time)
				newEvents = append(newEvents, event)

			} else if event.Time.Before(c.Open) || event.Time.After(c.Close) {
				event := createErrorEvent("NotOpenYet", event.Time)
				newEvents = append(newEvents, event)

			} else {
				c.Clients[event.Client] = 1
			}
		} else if event.ID == 2 {
			if _, ok := c.Clients[event.Client]; !ok {
				event := createErrorEvent("ClientUnknown", event.Time)
				newEvents = append(newEvents, event)

			} else if c.TablesStatus[event.Table] != "" {
				event := createErrorEvent("PlaceIsBusy", event.Time)
				newEvents = append(newEvents, event)

			} else {
				for i := 1; i <= c.Tables; i++ {
					if c.TablesStatus[i] == event.Client {
						// Если клиента пересаживается за тот же стол, за которым сидит
						if event.Table == i {
							event := createErrorEvent("AlredySittingOnThisTable", event.Time)
							newEvents = append(newEvents, event)
						} else {
							// При пересадке клиента за другой стол создаем костыльное событие, чтобы не сломать рассчет выручки для кждого стола
							c.TablesStatus[i] = ""
							event := Event{
								Time:   event.Time,
								ID:     14,
								Client: event.Client,
								Table:  i,
							}
							newEvents = append(newEvents, event)
						}
					}

				}
				c.TablesStatus[event.Table] = event.Client
			}
		} else if event.ID == 3 {
			if _, ok := c.Clients[event.Client]; !ok {
				event := createErrorEvent("ClientUnknown", event.Time)
				newEvents = append(newEvents, event)
			} else if event.Time.Before(c.Open) || event.Time.After(c.Close) {
				event := createErrorEvent("NotOpenYet", event.Time)
				newEvents = append(newEvents, event)
			} else if c.isTableFreeExists() {
				event := createErrorEvent("ICanWaitNoLonger!", event.Time)
				newEvents = append(newEvents, event)
			} else {
				if len(c.WaitingQueue) > c.Tables {
					event := Event{
						Time:   event.Time,
						ID:     11,
						Client: event.Client,
					}
					newEvents = append(newEvents, event)
				} else {
					c.WaitingQueue = append(c.WaitingQueue, event.Client)
				}
			}
		} else if event.ID == 4 {
			if _, ok := c.Clients[event.Client]; !ok {
				event := createErrorEvent("ClientUnknown", event.Time)
				newEvents = append(newEvents, event)
			} else {
				for i, table := range c.TablesStatus {
					if table == event.Client {
						c.TablesStatus[i] = ""
						break
					}
				}
				delete(c.Clients, event.Client)
				if len(c.WaitingQueue) > 0 {
					client := c.WaitingQueue[0]
					c.WaitingQueue = c.WaitingQueue[1:]
					c.Clients[client] = 0
					for i, table := range c.TablesStatus {
						if table == "" {
							c.TablesStatus[i] = client
							event := Event{
								Time:   event.Time,
								ID:     12,
								Client: client,
								Table:  i,
							}
							newEvents = append(newEvents, event)
							break
						}
					}
				}
			}
		} else if event.ID == 11 {
			if _, ok := c.Clients[event.Client]; !ok {
				event := createErrorEvent("ClientUnknown", event.Time)
				newEvents = append(newEvents, event)
			} else {
				delete(c.Clients, event.Client)
				for i, table := range c.TablesStatus {
					if table == event.Client {
						c.TablesStatus[i] = ""
						break
					}
				}
			}
		} else if event.ID == 12 {
			if _, ok := c.Clients[event.Client]; !ok {
				event := createErrorEvent("ClientUnknown", event.Time)
				newEvents = append(newEvents, event)
			} else {
				c.TablesStatus[event.Table] = event.Client
			}
		}
	}

	// Создание среза имен клиентов из мапы имен клиентов
	clients := make([]string, len(c.Clients))
	i := 0
	for k := range c.Clients {
		clients[i] = k
		i++
	}

	// Сортируем срез имен клиентов
	sort.Slice(clients, func(i, j int) bool {
		return clients[i] < clients[j]
	})

	// После закрытия клуба клиенты принудительно уходят
	for _, client := range clients {
		event := Event{
			Time:   c.Close,
			ID:     11,
			Client: client,
		}
		newEvents = append(newEvents, event)
	}

	// Выводим все ивенты, включая ошибки, в консоль
	for _, event := range newEvents {
		if event.ID != 14 {
			if event.ID != 13 {
				fmt.Printf("%s %d %s", event.Time.Format("15:04"), event.ID, event.Client)
				if event.Table != 0 {
					fmt.Printf(" %d\n", event.Table)
				} else {
					fmt.Print("\n")
				}
			} else {
				fmt.Printf("%s %d %s\n",
					event.Time.Format("15:04"),
					event.ID,
					event.Error,
				)
			}
		}
	}

	// Выводим время закрытия компьютерного клуба
	fmt.Println(c.Close.Format("15:04"))

	// Расчет прибыли для каждого стола отдельно
	for i := 1; i <= c.Tables; i++ {
		income := 0
		prevClient := ""
		prevTime := time.Time{}
		var busyTime time.Duration
		var totalBusyTime time.Duration
		time := time.Time{}

		// Вычисление прибыли конкретного стола
		for k, event := range newEvents {
			if event.Table == i || event.ID == 4 || event.ID == 11 {
				if event.ID == 2 && newEvents[k+1].ID != 13 || event.ID == 12 {
					prevClient = event.Client
					prevTime = event.Time
				} else if event.ID == 4 || event.ID == 11 || event.ID == 14 {
					if prevClient == event.Client {
						busyTime = event.Time.Sub(prevTime)
						totalBusyTime += busyTime
						hours := int(math.Ceil(busyTime.Hours()))
						income += hours * c.Price
						busyTime = 0
						prevClient = ""
					}
				}
			}
		}
		time = time.Add(totalBusyTime)
		fmt.Printf("%d %d %s\n", i, income, time.Format("15:04"))
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: task <filename>")
		os.Exit(1)
	}
	club := &Club{}
	if err := club.parseClub(os.Args[1]); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
	club.ProcessEvents()
}
