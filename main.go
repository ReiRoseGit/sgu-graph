package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/cheekybits/genny/generic"
)

// Generic для описания узла
type Item generic.Type

type Node struct {
	value Item
}

// toString - функция для строчного представления узла
func (n *Node) toString() string {
	return fmt.Sprintf("%v", n.value)
}

/*
Graph - конструктор, создающий пустой граф:
- mutex - блокировка структуры
- is_oriented - ориентированный ли граф
- is_suspended - взвешенный ли граф
- edges - ребра / дуги графа
*/
type Graph struct {
	mutex        sync.Mutex
	is_oriented  bool
	is_suspended bool
	edges        map[*Node]map[*Node]int
}

/*

Конструкторы:
- newEmptyGraph - конструктор, возвращающий пустой, неориентированный, взвешенный граф
- newCopiedGraph - функция для глубокого копирования графа, возвращает ссылку на свою полную копию
- newGraphFromFile - возвращает граф, созданный из данный файла
- newCompleteGraph - создает полный граф, содержащий count вершин
*/

// newEmptyGraph - конструктор, возвращающий пустой, ориентированный, взвешенный граф
func newEmptyGraph() *Graph {
	return &Graph{sync.Mutex{}, true, true, make(map[*Node]map[*Node]int)}
}

// newCopiedGraph - функция для глубокого копирования графа, возвращает ссылку на свою полную копию
func newCopiedGraph(g *Graph) *Graph {
	newGraph := newEmptyGraph()
	newGraph.is_oriented = g.is_oriented
	newGraph.is_suspended = g.is_suspended
	newGraph.edges = make(map[*Node]map[*Node]int, len(g.edges))
	for k1, v1 := range g.edges {
		for k2, v2 := range v1 {
			newGraph.addEdge(k1.toString(), k2.toString(), v2)
		}
	}
	return newGraph
}

// newGraphFromFile - возвращает граф, созданный из данных файла.
// Если при выполнении возникает ошибка, то возвращает ее и пустой граф
func newGraphFromFile(path string) (*Graph, error) {
	g := newEmptyGraph()
	data, err := getDataFromFile(path)
	if err != nil {
		return g, err // Пустой граф и ошибка
	}

	// Ориентированность
	if data[0] == "unoriented" {
		g.is_oriented = false
	}

	// Взвешенность
	if data[1] == "unsuspended" {
		g.is_suspended = false
	}

	// Заполнени узлов и дуг / ребер
	for i := 2; i < len(data); i++ {
		currentData := strings.Split(data[i], " ")
		currentDistance, err := strconv.Atoi(currentData[2])
		if err != nil {
			return g, err // Ошибка преобразования числа
		}
		g.addEdge(currentData[0], currentData[1], currentDistance)
	}

	return g, nil
}

// newCompleteGraph - создает полный граф, содержащий count вершин.
// Граф является неориентированный, невзвешенным и не содержит петель
func newCompleteGraph(count int) *Graph {
	g := newEmptyGraph()
	g.is_suspended = false
	g.is_oriented = false
	names := []string{}
	for i := 0; i < count; i++ {
		var name string
		fmt.Println("Введите:", i+1, "название:")
		fmt.Scan(&name)
		names = append(names, name)
	}
	for i := 0; i < len(names); i++ {
		for j := 1; j < len(names); j++ {
			if i == j {
				continue
			}
			g.addEdge(names[i], names[j], 0)
		}
	}
	return g
}

/*

Методы:
- addNode - добавляет вершину в граф
- addEdge - добавляет дугу / ребро между узлами
- removeEdge - удаляет дугу / ребро
- removeNode - удаляет узел и все входящие и исходящие ребра / дуги
- printDataInFile - выводит данные о графе в файл

*/

// addNode - добавляет вершину в граф
func (g *Graph) addNode(value string) *Node {
	ref := g.getRefOfNode(value)
	if ref == nil {
		node := &Node{value}
		g.edges[node] = map[*Node]int{}
		return node
	} else {
		return ref
	}
}

// addEdge - добавляет дугу / ребро между узлами,
// если соединить два узла, между которыми уже есть связь, то перезапишет ее.
// Если узла нет, то создаст его
func (g *Graph) addEdge(value1, value2 string, distance int) {
	ref1 := g.addNode(value1)
	ref2 := g.addNode(value2)
	if !g.is_oriented {
		if g.is_suspended {
			g.edges[ref1][ref2] = distance
			g.edges[ref2][ref1] = distance
		} else {
			g.edges[ref1][ref2] = -1
			g.edges[ref2][ref1] = -1
		}
	} else {
		if g.is_suspended {
			g.edges[ref1][ref2] = distance
		} else {
			g.edges[ref1][ref2] = -1
		}
	}
}

// removeEdge - удаляет дугу / ребро, если какого-то элемента не существует,
// то ничего не удаляет
func (g *Graph) removeEdge(value1, value2 string) {
	node1 := g.getRefOfNode(value1)
	node2 := g.getRefOfNode(value2)
	if node1 != nil && node2 != nil {
		if !g.is_oriented {
			delete(g.edges[node1], node2)
			delete(g.edges[node2], node1)
		} else {
			delete(g.edges[node1], node2)
		}
	} else {
		fmt.Println("Оба узла должны существовать в графе", value1, value2)
	}
}

// removeNode - удаляет узел и все входящие и исходящие ребра / дуги,
// если узла не существует, то возвращает ошибку
func (g *Graph) removeNode(value string) {
	node := g.getRefOfNode(value)
	if node != nil {
		for k := range g.edges {
			g.removeEdge(k.toString(), value)
			g.removeEdge(value, k.toString())
		}
		delete(g.edges, node)
	} else {
		fmt.Println("Узел должен существовать в графе")
	}
}

// printDataInFile - выводит данные о графе в файл, данные пригодны для создания нового графа
// с помощью newGraphFromFile
func (g *Graph) printDataInFile(path string) error {
	file, err := os.Create(path)
	if g.is_oriented {
		file.WriteString("oriented\n")
	} else {
		file.WriteString("unoriented\n")
	}
	if g.is_suspended {
		file.WriteString("suspended\n")
	} else {
		file.WriteString("unsuspended\n")
	}
	for k := range g.edges {
		for k2, v2 := range g.edges[k] {
			var currentString string
			if g.is_suspended {
				currentString = fmt.Sprintf("%s %s %d\n", k.toString(), k2.toString(), v2)
			} else {
				currentString = fmt.Sprintf("%s %s %d\n", k.toString(), k2.toString(), -1)
			}
			file.WriteString(currentString)
		}
	}
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

/*

Вспомогательные функции, необходимые для выполнения задания:

*/

// getDataFromFile - функция для считывания данных из файла
func getDataFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error when opening file: %s", err)
		return make([]string, 0), err
	}
	fileScanner := bufio.NewScanner(file)
	var result []string
	for fileScanner.Scan() {
		result = append(result, fileScanner.Text())
	}
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error while reading file: %s", err)
		return make([]string, 0), err
	}
	file.Close()
	err = validateData(result)
	if err != nil {
		log.Fatalf(err.Error())
		return make([]string, 0), err
	}
	return result, nil
}

// validateData - проверка входных данных из файла
func validateData(str []string) error {
	if !(str[0] == "oriented" || str[0] == "unoriented") {
		log.Fatalf("Неправильный тип ориентации графа")
		return errors.New("Неправильный тип ориентации графа")
	}

	if !(str[1] == "suspended" || str[1] == "unsuspended") {
		log.Fatalf("Неправильный тип взвешенности графа")
		return errors.New("Неправильный тип взвешенности графа")
	}
	return nil
}

// getRefOfNode - возвращает ссылку на узел или nil
func (g *Graph) getRefOfNode(value string) *Node {
	for k := range g.edges {
		if k.toString() == value {
			return k
		}
	}
	return nil
}

// printNodes - выводит все узлы в графе
func (g *Graph) printNodes() {
	for k := range g.edges {
		fmt.Println("Узел:", k.toString())
	}
}

// printEdges - выводит узлы и их связи
func (g *Graph) printEdges() {
	for k, v := range g.edges {
		for k2, v2 := range v {
			if g.is_suspended {
				fmt.Println(k.toString(), "->", k2.toString(), ":", v2)
			} else {
				fmt.Println(k.toString(), "->", k2.toString())
			}
		}
	}
}

// Выводит узлы и связи в комфортном виде
func (g *Graph) printEdgesComfort() {
	for k, v := range g.edges {
		fmt.Println(k.toString() + ":")
		for k2, v2 := range v {
			if g.is_suspended {
				fmt.Println("\t", k2.toString(), ":", v2)
			} else {
				fmt.Println("\t", k2.toString())
			}
		}
	}
}

// printInformationAboutGraph - выводит всю информацию о графе
func (g *Graph) printInformationAboutGraph() {
	fmt.Println("Граф:")
	if g.is_oriented {
		fmt.Println("- Ориентированный")
	} else {
		fmt.Println("- Неориентированный")
	}
	if g.is_suspended {
		fmt.Println("- Взвешенный")
	} else {
		fmt.Println("- Невзвешенный")
	}
	fmt.Println("Узлы:")
	g.printNodes()
	fmt.Println("Связи в графе:")
	g.printEdgesComfort()
	fmt.Println("===============")
}

/*

Реализация консольного интерфейса:

*/

// chooseNextAction - позволяет пользователю выбрать следующее действие
func chooseNextAction() (string, error) {
	var input string
	fmt.Println("Введите действие:")
	fmt.Println("Конструкторы:")
	fmt.Println("1 - Создать пустой граф;")
	fmt.Println("2 - Ввести данные из файла;")
	fmt.Println("3 - Создать копию графа;")
	fmt.Println("4 - Создать полный граф n вершин;")
	fmt.Println("Методы:")
	fmt.Println("5 - Добавить узел;")
	fmt.Println("6 - Добавить ребро / дугу;")
	fmt.Println("7 - Удалить узел;")
	fmt.Println("8 - удалить ребро / дугу;")
	fmt.Println("9 - Вывести данные в файл;")
	fmt.Println("10 - Распечатать информацию о графе в консоль;")
	fmt.Println("11 - Вывести полустепень захода;")
	fmt.Println("12 - Вывести все узлы орграфа не смежные с данным;")
	fmt.Println("13 - Построить новый граф, удалив из него все вершины с нечеными степенями;")
	fmt.Println("14 - Найти путь, соединяющий вершины u1 и u2 и не проходящий через вершину v;")
	fmt.Println("15 - Проверить является ли граф деревом / лесом;")
	fmt.Println("16 - Выполнить обход графа в глубину;")
	fmt.Println("17 - Выполнить обход графа в ширину;")
	fmt.Println("18 - Алгоритм Прима;")
	fmt.Println("19 - Алгоритм Дейкстры в чистом виде;")
	fmt.Println("20 - Алгоритм Дейкстры - найти радиус графа;")
	fmt.Println("21 - Алгоритм Флойда в чистом виде - кратчайшие пути между всеми парами вершин;")
	fmt.Println("22 - Алгоритм Беллмана - найти кратчайший путь между заданной парой вершин;")
	fmt.Println("23 - Найти максимальный поток в графе;")
	fmt.Println("0 - Остановить выполнение программы;")
	fmt.Scan(&input)
	err := validateAction(input)
	if err != nil {
		return "", err
	}
	return input, nil
}

// validateAction - выполняет проверку правильности данных, введеных пользователем
func validateAction(action string) error {
	value, err := strconv.Atoi(action)
	if err != nil {
		fmt.Println("Неккоректная операция, введите число от 0 до 23")
		return errors.New("Неккоректная операция")
	}
	if value < 0 || value > 23 {
		fmt.Println("Неккоректная операция, введите число от 0 до 23")
		return errors.New("Неккоректная операция")
	}
	return nil
}

// validateNode - проверяет вершину графа на существование
func validateNode(g *Graph, value string) error {
	node := g.getRefOfNode(value)
	if node == nil {
		return errors.New("Вершина " + value + " не существует в графе!")
	}
	return nil
}

// validateDistance - проверяет корректность введенного расстояния
func validateDistance(value string) (int, error) {
	d, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, errors.New("Некорректное значение расстояния")
	}
	return d, nil
}

// Реализация консольного интерфейса
func consoleInterface() {
	var workingGraph *Graph
	workingGraph = nil
act:
	for {
		action, err := chooseNextAction()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		a, err := strconv.Atoi(action)
		if workingGraph == nil && a > 4 {
			fmt.Println("Граф не задан! Задайте граф и повторите попытку.")
			continue
		}
		// Выбор действия
		switch action {
		case "0":
			fmt.Println("Выполнение программы остановлено!")
			break act
		case "1":
			workingGraph = newEmptyGraph()
		case "2":
			var path string
			fmt.Println("Введите путь к файлу:")
			fmt.Scan(&path)
			workingGraph, err = newGraphFromFile(path)
			if err != nil {
				fmt.Println("Произошла ошибка")
				fmt.Println(err.Error())
				break act
			}
		case "3":
			workingGraph = newCopiedGraph(workingGraph)
		case "4":
			var count string
			fmt.Println("Введите количество вершин:")
			fmt.Scan(&count)
			c, err := strconv.Atoi(count)
			if err != nil {
				fmt.Println("Произошла ошибка!")
				fmt.Println(err.Error())
				break act
			}
			workingGraph = newCompleteGraph(c)
		case "5":
			var node string
			fmt.Println("Введите узел:")
			fmt.Scan(&node)
			check := workingGraph.getRefOfNode(node)
			if check != nil {
				fmt.Println("Вершина", node, "уже существует!")
				continue
			}
			workingGraph.addNode(node)
		case "6":
			var node1, node2 string
			fmt.Println("Введите узел 1:")
			fmt.Scan(&node1)
			fmt.Println("Введите узел 2:")
			fmt.Scan(&node2)
			err = validateNode(workingGraph, node1)
			if err != nil {
				fmt.Println("Произошла ошибка!")
				fmt.Println(err.Error())
				continue
			}
			err = validateNode(workingGraph, node2)
			if err != nil {
				fmt.Println("Произошла ошибка!")
				fmt.Println(err.Error())
				continue
			}
			if workingGraph.is_suspended {
				var distance string
				fmt.Println("Введите расстояние: ")
				fmt.Scan(&distance)
				dist, err := validateDistance(distance)
				if err != nil {
					fmt.Println("Произошла ошибка!")
					fmt.Println(err.Error())
					continue
				}
				workingGraph.addEdge(node1, node2, dist)
			} else {
				workingGraph.addEdge(node1, node2, -1)
			}
		case "7":
			var node string
			fmt.Println("Введите узел:")
			fmt.Scan(&node)
			check := workingGraph.getRefOfNode(node)
			if check == nil {
				fmt.Println("Вершина не существует в графе!")
				continue
			}
			workingGraph.removeNode(node)
		case "8":
			var node1, node2 string
			fmt.Println("Введите узел 1:")
			fmt.Scan(&node1)
			fmt.Println("Введите узел 2:")
			fmt.Scan(&node2)
			check := workingGraph.getRefOfNode(node1)
			if check == nil {
				fmt.Println("Вершина", node1, "не существует в графе!")
				continue
			}
			check = workingGraph.getRefOfNode(node2)
			if check == nil {
				fmt.Println("Вершина", node2, "не существует в графе!")
				continue
			}
			workingGraph.removeEdge(node1, node2)
		case "9":
			var path string
			fmt.Println("Введите путь к файлу:")
			fmt.Scan(&path)
			workingGraph.printDataInFile(path)
		case "10":
			workingGraph.printInformationAboutGraph()
		case "11":
			var node string
			fmt.Println("Введите вершину:")
			fmt.Scan(&node)
			res := workingGraph.getInclinationDegree(node)
			if res != -1 {
				fmt.Println("Степень полузахода вершины", node, "равна:", res)
			}
		case "12":
			var node string
			fmt.Println("Введите вершину:")
			fmt.Scan(&node)
			workingGraph.printAllNonContiguousNodes(node)
		case "13":
			workingGraph = workingGraph.getNewGraphWithoutOddNodes()
		case "14":
			var node1, node2, node3 string
			fmt.Println("Введите вершину 1:")
			fmt.Scan(&node1)
			fmt.Println("Введите вершину 2:")
			fmt.Scan(&node2)
			fmt.Println("Введите вершину 3:")
			fmt.Scan(&node3)
			if workingGraph.getRefOfNode(node1) != nil && workingGraph.getRefOfNode(node2) != nil && workingGraph.getRefOfNode(node3) != nil {
				workingGraph.getCurrentWay(node1, node2, node3)
			} else {
				fmt.Println("Не все вершины существуют в графе")
			}
		case "15":
			workingGraph.isGraphTreeOrForest()
		case "16":
			var node string
			fmt.Println("Введите вершину:")
			fmt.Scan(&node)
			if workingGraph.getRefOfNode(node) != nil {
				workingGraph.Dfs(node)
			} else {
				fmt.Println("Вершина не существует в графе")
			}
		case "17":
			var node string
			fmt.Println("Введите вершину:")
			fmt.Scan(&node)
			if workingGraph.getRefOfNode(node) != nil {
				workingGraph.Bfs(node, true)
			} else {
				fmt.Println("Вершина не существует в графе")
			}
		case "18":
			var node string
			fmt.Println("Введите вершину:")
			fmt.Scan(&node)
			if workingGraph.getRefOfNode(node) != nil {
				workingGraph.prim(node)
			} else {
				fmt.Println("Вершина не существует в графе")
			}
		case "19":
			var node string
			fmt.Println("Введите вершину:")
			fmt.Scan(&node)
			ref := workingGraph.getRefOfNode(node)
			if ref != nil {
				workingGraph.Deikstra(ref, true)
			} else {
				fmt.Println("Вершина не существует в графе")
			}
		case "20":
			workingGraph.getRadiusOfGraph()
		case "21":
			workingGraph.Floyd()
		case "22":
			var node1, node2 string
			fmt.Println("Введите вершину u:")
			fmt.Scan(&node1)
			ref1 := workingGraph.getRefOfNode(node1)

			fmt.Println("Введите вершину v:")
			fmt.Scan(&node2)
			ref2 := workingGraph.getRefOfNode(node2)
			if ref1 != nil && ref2 != nil {
				workingGraph.Bellman(ref1, ref2, true)
			} else {
				fmt.Println("Вершины не существуют в графе!")
			}
		case "23":
			var node1, node2 string
			fmt.Println("Введите источник:")
			fmt.Scan(&node1)
			fmt.Println("Введите сток:")
			fmt.Scan(&node2)

			ref1 := workingGraph.getRefOfNode(node1)
			ref2 := workingGraph.getRefOfNode(node2)

			if ref1 != nil && ref2 != nil {
				workingGraph.getMaxFlow(ref1, ref2)
			} else {
				fmt.Println("Вершины не существуют в графе!")
			}
		}
	}
}

/*
Задачи:
Блок 1А:
4 - getInclinationDegree - выводит полустепень захода указанной вершины
20 - printAllNonContiguousNodes - выводит все вершины оргафа, не смежные с данной

Блок 1Б:
18 - getNewGraphWithoutOddNodes - возвращает новый граф без вершин с нечетными степенями

*/

// getInclinationDegree - выводит полустепень захода указанной вершины
func (g *Graph) getInclinationDegree(value string) int {
	node := g.getRefOfNode(value)
	if node == nil {
		fmt.Println("Узел не существует в графе")
		return -1
	}
	count := 0
	for key := range g.edges {
		if _, ok := g.edges[key][node]; ok {
			count++
		}
	}
	return count
}

// printAllNonContiguousNodes - выводит все вершины оргафа, не смежные с данной
func (g *Graph) printAllNonContiguousNodes(value string) {
	node := g.getRefOfNode(value)
	if node == nil {
		fmt.Println("Узел не существует в графе")
		return
	}
	if !g.is_oriented {
		fmt.Println("Граф является неориентированным")
		return
	}
	fmt.Println("Не смежные вершины:")
	counter := 0
	for key := range g.edges {
		isContiguous := false
		// эту проверку можно по идее убрать
		if key.toString() != value {
			if _, ok := g.edges[key][node]; ok {
				isContiguous = true
			}
			if _, ok := g.edges[node][key]; ok {
				isContiguous = true
			}
			if !isContiguous {
				counter++
				fmt.Println(key.toString())
			}
		}
	}
	if counter == 0 {
		fmt.Println("Все вершины графа смежны с данной")
	}

}

// getNewGraphWithoutOddNodes - возвращает граф, построенный однократным удалением вершин с нечетными степенями
func (g *Graph) getNewGraphWithoutOddNodes() *Graph {
	newG := newCopiedGraph(g)
	oddNodes := []*Node{}
	// создаем срез нечетных вершин графа
	for k := range g.edges {
		if g.getDegree(k.toString())%2 != 0 {
			oddNodes = append(oddNodes, k)
		}
	}
	// удаляем вершины
	for _, node := range oddNodes {
		newG.removeNode(node.toString())
	}
	return newG
}

// getCurrentWay - получает путь из u1 в u2, не проходящий через v
func (g *Graph) getCurrentWay(u1, u2, v string) {
	workingGraph := newCopiedGraph(g)
	node1 := workingGraph.getRefOfNode(u1)
	node2 := workingGraph.getRefOfNode(u2)
	node3 := workingGraph.getRefOfNode(v)
	if node1 == nil || node2 == nil || node3 == nil {
		fmt.Println("Не все узлы существуют в графе")
	}
	workingGraph.removeNode(v)
	queue := []*Node{node1}
	visited := []*Node{node1}
	way := make(map[string]string)
	isFindResult := false
	for e := range workingGraph.edges {
		way[e.toString()] = ""
	}
	for {
		if len(queue) == 0 {
			break
		}
		currentElement := queue[0]
		if currentElement.toString() == u2 {
			isFindResult = true
		}
		queue = queue[1:]
		for element := range workingGraph.edges[currentElement] {
			isVisited := false
			for _, v := range visited {
				if v == element {
					isVisited = true
				}
			}
			if !isVisited {
				queue = append(queue, element)
				visited = append(visited, element)
				way[element.toString()] = currentElement.toString()
			}
		}
	}
	if !isFindResult {
		fmt.Println("Пути не существует")
		return
	}
	answer := []string{u2}
	element := u2
	for {
		if element == u1 {
			break
		}
		element = way[element]
		answer = append(answer, element)
	}
	for ind := range answer {
		fmt.Println(answer[len(answer)-ind-1])
	}
}

// isGraphTreeOrForest - проверяет граф на дерево или лес
func (g *Graph) isGraphTreeOrForest() {
	countOfComponents := false // показатель того, что не 1 компонента связности
	count := len(g.edges)
	// цикл для всех вершин
	for n := range g.edges {
		visited := []string{n.toString()} // список посещений
		isNoCycle := true                 // нет ли цикла
		workingGraph := newCopiedGraph(g) // копия графа для удаления ребер
		nodeValue := workingGraph.getRefOfNode(n.toString())
		c := 0 // кол-во ребер
		workingGraph.dfsForTree(nodeValue, nodeValue, &visited, &isNoCycle, &c)
		if isNoCycle && len(visited) == c+1 { // проверка на дерево: n = m +1
			count -= 1
			if len(g.edges) != c+1 {
				countOfComponents = true // компонент связности более 1
			}
		}
	}
	// Если 1 компонента связности
	if !countOfComponents {
		if count == 0 {
			fmt.Println("Граф является деревом")
		} else {
			fmt.Println("Граф не является ни деревом, ни лесом")
		}
	} else {
		if count == 0 {
			fmt.Println("Граф является лесом")
		} else {
			fmt.Println("Граф не яввляется ни деревом, ни лесом")
		}
	}
}

// dfsForTree - выполняет обход графа в глубину, проверяя компоненту на "дерево"
func (g *Graph) dfsForTree(v, currentV *Node, visited *[]string, isNoCycle *bool, count *int) {
	if len(*visited) == len(g.edges) {
		return
	}
	// есть цикл или петля
	if _, ok := g.edges[currentV][v]; ok {
		fmt.Println("Связь:", currentV.toString(), v.toString())
		*isNoCycle = false
	}

	// проссматриваем всех соседей до первого непройденного
	for n := range g.edges[currentV] {
		// проверяем, что не заходили в узел
		isVisited := false
		for _, t := range *visited {
			if t == n.toString() {
				isVisited = true
			}
		}
		// если не заходили
		if !isVisited {
			*visited = append(*visited, n.toString())
			g.removeEdge(currentV.toString(), n.toString())
			// есть ли связь с уже проссмотренными
			for _, t := range *visited {
				if _, ok := g.edges[n][g.getRefOfNode(t)]; ok {
					*isNoCycle = false
				}
			}
			*count += 1
			g.dfsForTree(v, n, visited, isNoCycle, count)
		}
	}
}

// Bfs - выполняет обход графа в глубину, начиная с указанной вершины
// isPrintNeeded - указатель того нужен вывод в консоль или нет
func (g *Graph) Bfs(v string, isPrintNeeded bool) []string {
	visited := []string{v} // список посещенных вершин
	queue := []string{v}   // очередь для посещения
	for {
		// в очереди нет элементов
		if len(queue) == 0 {
			break
		}
		currentElement := queue[0]
		if isPrintNeeded {
			fmt.Println("Узел", currentElement)
		}
		queue = queue[1:]
		// Цикл по всем связям
		for element := range g.edges[g.getRefOfNode(currentElement)] {
			// Проверка на посещение
			isVisited := false
			for _, v := range visited {
				if v == element.toString() {
					isVisited = true
				}
			}
			// Если не посещали данный узел
			if !isVisited {
				visited = append(visited, element.toString())
				queue = append(queue, element.toString())
			}
		}
	}
	return visited
}

// Dfs - Выполняет обход графа в глубину, начиная с указанной вершины
func (g *Graph) Dfs(v string) {
	node := g.getRefOfNode(v)
	visited := []Node{*node} // посещенные вершины
	g.dfsHelper(node, &visited)
}

// dfsHelper - вспомогательная функция для обхода графа в глубину
func (g *Graph) dfsHelper(node *Node, visited *[]Node) {
	fmt.Println("Узел", node.toString())
	for nextNode := range g.edges[node] {
		isVisited := false
		for _, t := range *visited {
			if t.toString() == nextNode.toString() {
				isVisited = true
			}
		}
		if !isVisited {
			*visited = append(*visited, *nextNode)
			g.dfsHelper(nextNode, visited)
		}
	}
}

// prim - реализация алгоритма Прима
func (g *Graph) prim(v string) *Graph {
	result := newEmptyGraph()
	result.is_oriented = false
	result.is_suspended = true
	// Проверка графа на связность
	if len(g.Bfs(v, false)) != len(g.edges) {
		fmt.Println("Граф является несвязным!")
		return nil
	}
	visited := []*Node{g.getRefOfNode(v)} // список посещенных
	r := []string{}
	for n := range g.edges {
		if n.toString() != v {
			weight, element, parent := g.searchMin(visited)
			visited = append(visited, element)
			r = append(r, parent+" -> "+element.toString()+": "+strconv.Itoa(weight))
			result.addEdge(parent, element.toString(), weight)
		}
	}
	for _, t := range r {
		fmt.Println(t)
	}
	return result
}

// searchMin - ищет минимальный вес ребра
func (g *Graph) searchMin(visited []*Node) (int, *Node, string) {
	min := -1
	var index2 *Node
	var parent string
	// ищем максимальное значение среди всего списка смежности
	for e := range g.edges {
		for e2 := range g.edges[e] {
			if g.edges[e][e2] > min {
				min = g.edges[e][e2]
			}
		}
	}
	// выбираем минимальный вес, где один конец ребра принадлежит уже проссмотренным, а другой - нет
	for _, t := range visited {
		for elem, w := range g.edges[t] {
			// Проверка на посещенность
			isVisited := false
			for _, k := range visited {
				if elem.toString() == k.toString() {
					isVisited = true
				}
			}
			if !isVisited && w < min {
				min = w
				index2 = elem
				parent = t.toString()
			}
		}
	}
	return min, index2, parent
}

// Floyd - реализация алгоритма Флойда
func (g *Graph) Floyd() {

	// Составляем матрицу
	res := make(map[string]map[string]int)

	// Составляем пути
	path := make(map[string]map[string]string)

	// Заполняем матрицу
	for node1 := range g.edges {
		res[node1.toString()] = map[string]int{}
		for node2 := range g.edges {
			// Если между node1 и node2 есть ребро, его и запоминаем
			if distance, ok := g.edges[node1][node2]; ok && node1.toString() != node2.toString() {
				res[node1.toString()][node2.toString()] = distance
			} else {
				res[node1.toString()][node2.toString()] = 10000 // Иначе присваиваем недостижимое значение
			}
		}
	}

	// Заполняем список путей
	// 0 - дуга, -1 - пути не существует, узел1 - путь существует
	for node1 := range res {
		path[node1] = make(map[string]string)
		for node2 := range res {
			if node1 == node2 {
				path[node1][node2] = "0"
			} else {
				if res[node1][node2] != 10000 {
					path[node1][node2] = node1
				} else {
					path[node1][node2] = "-1"
				}
			}
		}
	}

	// Сам алгоритм
	// Внешний цикл по всем вершинам графа
	for n1 := range res {
		// Просматриваем строчку I
		for n2 := range res {
			// Просматриваем строчку II
			for n3 := range res {
				// Формируем новое расстояние
				// Задаемся вопросом: быстрее ли пройти через внешнюю вершину или напрямую
				if res[n2][n1] < 10000 && res[n1][n3] < 10000 {
					res[n2][n3] = min(res[n2][n3], res[n2][n1]+res[n1][n3])
					if res[n2][n3] == res[n2][n1]+res[n1][n3] {
						path[n2][n3] = path[n1][n3]
					}
				}
			}
		}
	}

	// Выводим все кратчайшие расстояния
	fmt.Println("Кратчайшие пути между всеми парами вершин:")
	for n, v := range res {
		for t, d := range v {
			if n != t && path[n][t] != "-1" {
				fmt.Println("Кратчайшее расстояние между", n, "и", t, "составляет", d, "путь:")
				fmt.Print(n, " -> ")
				g.printPath(path, n, t)
				fmt.Print(t)
				fmt.Println()
			}
		}
	}
}

// printPath - вывод пути от одной вершины до другой
func (g *Graph) printPath(path map[string]map[string]string, n, t string) {
	if path[n][t] == n {
		return
	}
	g.printPath(path, n, path[n][t])
	fmt.Print(path[n][t], " -> ")
}

// min - возвращает минимальное значение из двух переданных параметров
func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

// Алгоритм Дейкстры - находит минимальные пути от вершины до всех остальных
func (g *Graph) Deikstra(beginNode *Node, isNeedOutput bool) map[*Node]int {

	// Минимальные расстояния от источника до вершин
	distances := make(map[*Node]int)

	// Посещенные вершины
	visited := make(map[*Node]bool)

	// Заполняем все вершины как непосещенные и присваим им недостижимое расстояние
	for n := range g.edges {
		distances[n] = 10000
		visited[n] = false
	}

	// Начальная вершина имеет метку 0, от нее до самой себя расстояние 0
	distances[beginNode] = 0
	for {
		var minIndex *Node // ближайшая вершина
		minIndex = nil
		min := 10000 // расстояние до ближайшей вершины

		// Ищем ближайшую непосещенную вершину
		for n := range g.edges {
			if distances[n] < min && !visited[n] {
				min = distances[n] // расстояние до этой вершины от источника
				minIndex = n       // сама вершина
			}
		}

		// Если нашли ближайшую непосещенную вершину
		if minIndex != nil {
			// проссматриваем все вершины, достижимые из найденной
			for n := range g.edges {
				// Проверка на положительное расстояние
				if g.edges[minIndex][n] > 0 {
					// минимальное расстояние до новой вершины формируется, как сумма расстояния от источника до текущей
					// и длины ребра от текущей до новой
					temp := min + g.edges[minIndex][n]
					// Если удалось укоротить расстояние
					if temp < distances[n] {
						distances[n] = temp
					}
				}
			}
			// вершина проссмотрена
			visited[minIndex] = true
		} else {
			break // Если нет вершин для рассмотрения
		}
	}

	// Вывод всех кратчайших расстояний от источника до остальных вершин (опционально)
	if isNeedOutput {
		fmt.Println("Кратчайшие расстояние от вершины:", beginNode.toString())
		for n, d := range distances {
			fmt.Println("Минимальное расстояние от вершины:", beginNode.toString(), "до вершины:", n.toString(), "равно:", d)
		}
	}
	// возвращаем словарь: ребро -> кратчайшего расстояние от источника до него
	return distances
}

// getRadiusOfGraph - находит радиус графа - минимальный из эксцентриситетов
func (g *Graph) getRadiusOfGraph() {
	// список максимальных расстояний
	maxDistances := []int{}
	for n := range g.edges {
		r := g.Deikstra(n, true) // Находим минимальные расстояния от всех вершни до текущей

		currentMax := -1
		// Находим максимальное из таких (минимальных от всех вершин до данной) расстояний
		for t, v := range r {
			if t.toString() != n.toString() {
				if currentMax < v {
					currentMax = v
				}
			}
		}

		// Сохраняем максимальное значение
		maxDistances = append(maxDistances, currentMax)
	}

	// Находим радиус - минимум из максимумов
	minDistance := 10000
	for _, v := range maxDistances {
		if v < minDistance {
			minDistance = v
		}
	}
	fmt.Println("Радиус графа равен:", minDistance)
}

// Bellman - алгоритм Беллмана, в данной реализации позволяет вывести кратчайшее растояние между двумя вершинами
func (g *Graph) Bellman(n, answerNode *Node, isNeedOutput bool) {
	// Словарь расстояний
	res := make(map[*Node]int)

	for n := range g.edges {
		res[n] = 100000
	}

	// Текущая вершина имеет расстояние 0
	res[n] = 0

	path := make(map[string]string)
	for n := range g.edges {
		path[n.toString()] = n.toString()
	}

	// Нужно выполнить n - 1 итерацию
	for i := 0; i < len(res)-1; i++ {
		// Проссматриваем все вершины и вычисляем кратчайшие расстояния
		for n, v := range g.edges {
			for t, d := range v {
				// Если расстояние от источника до рассматриваемой больше, чем сумма
				// расстояний от текущей до рассматриваемой + расстояние от текущей до итерируемой
				// то обновляем расстояние
				if res[n] != 100000 && res[n]+d < res[t] {
					res[t] = res[n] + d
					path[t.toString()] = n.toString()
				}
			}
		}
	}

	// Проверка на отрицательные циклы
	// выполняем еще один шаг и, если удалось укоротить расстояние => есть отрицательный цикл
	for n, v := range g.edges {
		for t, d := range v {
			if n.toString() != t.toString() {
				if res[n] != 100000 && res[n]+d < res[t] {
					fmt.Println("В графе есть отрицательный цикл!")
					return
				}
			}
		}
	}

	// Если нужен вывод кратчайших путей до всех вершин (опционально)
	if isNeedOutput {
		for node, d := range res {
			if node.toString() != n.toString() {
				fmt.Println(n.toString(), "->", node.toString(), d)
			}
		}
	}

	// Если путь не найден или найден
	if res[answerNode] != 100000 {
		fmt.Println("Расстояние между", n.toString(), "и", answerNode.toString(), "составляет:", res[answerNode])
		fmt.Println("Путь:")
		current := path[answerNode.toString()]
		p := []string{answerNode.toString()}
		for {
			if current == n.toString() {
				p = append(p, n.toString())
				break
			} else {
				p = append(p, current)
				current = path[current]
			}
		}
		for i := len(p) - 1; i > -1; i-- {
			if i == 0 {
				fmt.Print(p[i])
				fmt.Println()
			} else {
				fmt.Print(p[i], " -> ")
			}
		}
	} else {
		fmt.Println("Кратчайшего пути не существует")
	}

}

/*
bfsHelper - обход в ширину. Возвращает кратчайшие пути (по сумме дуг) от s до каждой вершины графа.
Попутно изменяет значение потока из s в каждую вершину.
*/
func (g *Graph) bfsHelper(s, t *Node, C, F *map[*Node]map[*Node]int, push *map[*Node]int, pred *map[*Node]*Node) bool {
	queue := []*Node{}
	visited := make(map[*Node]bool)
	for node := range g.edges {
		visited[node] = false
	}
	queue = append(queue, s)
	visited[s] = true
	(*pred)[s] = s
	(*push)[s] = 100000

	for {
		if visited[t] || len(queue) == 0 {
			break
		}
		u := queue[0]
		queue = queue[1:]
		for edge := range g.edges {
			// Если не посещали и поток не превосходит пропускную способность
			if !visited[edge] && ((*C)[u][edge]-(*F)[u][edge] > 0) {
				visited[edge] = true
				queue = append(queue, edge)
				(*push)[edge] = min((*push)[u], (*C)[u][edge]-(*F)[u][edge])
				(*pred)[edge] = u
			}
		}
	}
	return visited[t]
}

// initPredAndPush - возвращает поток из начальной вершины в v
// и словарь, показывающий откуда пришли в v (предок)
func (g *Graph) initPredAndPush() (*map[*Node]int, *map[*Node]*Node) {
	// Формируем словарь потоков
	push := make(map[*Node]int)
	for edge := range g.edges {
		push[edge] = 0
	}
	// Формируем словарь предков
	pred := make(map[*Node]*Node)
	for edge := range g.edges {
		pred[edge] = nil
	}
	return &push, &pred
}

// initFlowsAndBandwidth - возвращаем словарь пропускных способностей и текущих потоков в графе
func (g *Graph) initFlowsAndBandwidth() (*map[*Node]map[*Node]int, *map[*Node]map[*Node]int) {
	C := make(map[*Node]map[*Node]int) // Пропускные способности каждой дуги
	F := make(map[*Node]map[*Node]int) // Текущий поток в графе

	for node1 := range g.edges {
		C[node1] = map[*Node]int{}
		F[node1] = map[*Node]int{}
		for node2 := range g.edges {
			F[node1][node2] = 0
			if d, ok := g.edges[node1][node2]; ok {
				C[node1][node2] = d
			} else {
				C[node1][node2] = 0
			}
		}
	}
	return &C, &F
}

// Алгорит Форда-Фалкерсона - поиск макисмального потока в графе
// s - исток, t - сток
func (g *Graph) getMaxFlow(s, t *Node) {
	var u, v *Node
	flow := 0

	// Инициализация проп. способностей и текущих потоков
	C, F := g.initFlowsAndBandwidth()

	for {
		// Инициализация потомков и предков
		push, pred := g.initPredAndPush()
		if !g.bfsHelper(s, t, C, F, push, pred) {
			break
		}
		add := (*push)[t]
		v = t
		u = (*pred)[v]

		for {
			if v == s {
				break
			}
			(*F)[u][v] += add
			(*F)[v][u] -= add
			v = u
			u = (*pred)[v]
		}
		flow += add
	}
	fmt.Println("Максимальный поток:", flow)
}

/*
Вспомогательные функции к задачам:
- getDegree - возвращает степень вершины

*/

// getDegree - возвращает степень вершины
func (g *Graph) getDegree(value string) int {
	node := g.getRefOfNode(value)
	if node == nil {
		fmt.Println("Узел не существует в графе")
		return -1
	}
	count := 0
	if g.is_oriented {
		for key := range g.edges {
			if _, ok := g.edges[key][node]; ok {
				count++
			}
			if _, ok := g.edges[node][key]; ok {
				count++
			}
		}
		// Если есть петля, то она была подсчитана два раза, нужно вычесть единицу
		if _, ok := g.edges[node][node]; ok {
			count--
		}
	} else {
		for key := range g.edges {
			if _, ok := g.edges[node][key]; ok {
				count++
			}
		}
	}
	return count
}

func main() {
	consoleInterface()
}
