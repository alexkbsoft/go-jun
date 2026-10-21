package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		fmt.Println("Ошибка: ожидается подкоманда (например, 'versions')")
		os.Exit(1)
	}
	versionsCmd := flag.NewFlagSet("versions", flag.ExitOnError)
	var versionsServices string
	versionsCmd.StringVar(&versionsServices, "s", "", "Список сервисов через запятую")

	deployedCmd := flag.NewFlagSet("is-deployed", flag.ExitOnError)
	var deployedServices string
	deployedCmd.StringVar(&deployedServices, "s", "", "Список сервисов через запятую")

	switch os.Args[1] {
	case "versions":
		versionsCmd.Parse(os.Args[2:])
		services, err := parseServices(versionsServices, versionsCmd)
		if err != nil {
			return err
		}

		return runVersionsCommand(services)

	case "is-deployed":
		deployedCmd.Parse(os.Args[2:])
		parseServices(deployedServices, deployedCmd)

		// TODO: Здесь можно добавить логику для проверки развернутых сервисов

	default:
		return fmt.Errorf("неизвестная подкоманда: %s", os.Args[1])
	}

	return nil
}

func parseServices(input string, fs *flag.FlagSet) ([]string, error) {
	if input == "" {
		fmt.Println("Ошибка: флаг -s обязателен")
		fs.Usage()
		return nil, fmt.Errorf("Ошибка: флаг -s обязателен")
	}

	// Разбиваем строку и очищаем элементы от случайных пробелов
	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result, nil
}

func runVersionsCommand(services []string) error {
	for _, service := range services {
		url := fmt.Sprintf("http://%s/debug/info", service)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("Ошибка при запросе версии для сервиса %s: %v\n", service, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Сервис %s вернул статус: %s\n", service, resp.Status)
			continue
		}

		var str1, str2, version string
		// Scan string Commit SHA1: dc40c44c58170378c213eca00eb7bcd5af0e5c22 for sha
		if _, err := fmt.Fscan(resp.Body, &str1, &str2, &version); err != nil {
			return fmt.Errorf("Ошибка при чтении версии для сервиса %s: %v\n", service, err)
		}

		fmt.Printf("Сервис: %s, Версия: %s\n", service, version)
	}

	return nil
}
