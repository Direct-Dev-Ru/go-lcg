.PHONY: build build-podman run run-podman stop stop-podman logs logs-podman clean help

# Переменные
IMAGE_NAME = lcg-ollama
IMAGE_TAG = latest
CONTAINER_NAME = lcg-ollama
DOCKERFILE = Dockerfile
CONTEXT = ../..

help: ## Показать справку
	@echo "Доступные команды:"
	@echo "  make build          - Собрать Docker образ"
	@echo "  make build-podman   - Собрать Podman образ"
	@echo "  make run            - Запустить контейнер (Docker)"
	@echo "  make run-podman     - Запустить контейнер (Podman)"
	@echo "  make stop           - Остановить контейнер (Docker)"
	@echo "  make stop-podman    - Остановить контейнер (Podman)"
	@echo "  make logs           - Показать логи (Docker)"
	@echo "  make logs-podman    - Показать логи (Podman)"
	@echo "  make clean          - Удалить контейнер и образ"
	@echo "  make compose-up     - Запустить через docker-compose"
	@echo "  make compose-down   - Остановить docker-compose"
	@echo "  make podman-compose-up   - Запустить через podman-compose"
	@echo "  make podman-compose-down - Остановить podman-compose"

build: ## Собрать Docker образ
	docker build -f $(DOCKERFILE) -t $(IMAGE_NAME):$(IMAGE_TAG) $(CONTEXT)
	@echo "Образ $(IMAGE_NAME):$(IMAGE_TAG) успешно собран"

build-podman: ## Собрать Podman образ
	podman build -f $(DOCKERFILE) -t $(IMAGE_NAME):$(IMAGE_TAG) $(CONTEXT)
	@echo "Образ $(IMAGE_NAME):$(IMAGE_TAG) успешно собран"

run: ## Запустить контейнер (Docker)
	docker run -d \
		--name $(CONTAINER_NAME) \
		-p 8080:8080 \
		-p 11434:11434 \
		-v ollama-data:/home/ollama/.ollama \
		-v lcg-results:/app/data/results \
		-v lcg-prompts:/app/data/prompts \
		-v lcg-config:/app/data/config \
		$(IMAGE_NAME):$(IMAGE_TAG)
	@echo "Контейнер $(CONTAINER_NAME) запущен"

run-podman: ## Запустить контейнер (Podman)
	podman run -d \
		--name $(CONTAINER_NAME) \
		-p 8080:8080 \
		-p 11434:11434 \
		-v ollama-data:/home/ollama/.ollama \
		-v lcg-results:/app/data/results \
		-v lcg-prompts:/app/data/prompts \
		-v lcg-config:/app/data/config \
		$(IMAGE_NAME):$(IMAGE_TAG)
	@echo "Контейнер $(CONTAINER_NAME) запущен"

stop: ## Остановить контейнер (Docker)
	docker stop $(CONTAINER_NAME) || true
	docker rm $(CONTAINER_NAME) || true
	@echo "Контейнер $(CONTAINER_NAME) остановлен и удален"

stop-podman: ## Остановить контейнер (Podman)
	podman stop $(CONTAINER_NAME) || true
	podman rm $(CONTAINER_NAME) || true
	@echo "Контейнер $(CONTAINER_NAME) остановлен и удален"

logs: ## Показать логи (Docker)
	docker logs -f $(CONTAINER_NAME)

logs-podman: ## Показать логи (Podman)
	podman logs -f $(CONTAINER_NAME)

clean: ## Удалить контейнер и образ
	docker stop $(CONTAINER_NAME) || true
	docker rm $(CONTAINER_NAME) || true
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG) || true
	@echo "Контейнер и образ удалены"

compose-up: ## Запустить через docker-compose
	docker-compose up -d
	@echo "Сервисы запущены через docker-compose"

compose-down: ## Остановить docker-compose
	docker-compose down
	@echo "Сервисы остановлены"

podman-compose-up: ## Запустить через podman-compose
	podman-compose -f podman-compose.yml up -d
	@echo "Сервисы запущены через podman-compose"

podman-compose-down: ## Остановить podman-compose
	podman-compose -f podman-compose.yml down
	@echo "Сервисы остановлены"

shell: ## Подключиться к контейнеру (Docker)
	docker exec -it $(CONTAINER_NAME) sh

shell-podman: ## Подключиться к контейнеру (Podman)
	podman exec -it $(CONTAINER_NAME) sh

pull-model: ## Загрузить модель codegeex4 (Docker)
	docker exec $(CONTAINER_NAME) ollama pull codegeex4

pull-model-podman: ## Загрузить модель codegeex4 (Podman)
	podman exec $(CONTAINER_NAME) ollama pull codegeex4

