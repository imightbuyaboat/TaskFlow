# TaskFlow

В данном репозитории представлен бэкенд-сервис для управления очередями задач с распределенной обработкой, написанный на языке Go. Пользователи могут создавать задачи, которые помещаются в очередь и обрабатываются воркерами. Система поддерживает отложенные задачи, повторы в случае неудачного выполнения задачи, а также отображает статус задач и историю выполнения.

## Технологический стек

1. `Go` - основной язык разработки;
2. `PostgreSQL` - основное хранилище данных (пользователи, задачи, история);
3. `RabbitMQ` - брокер сообщений между сервисами.

## Компоненты системы

1. `Task-API` - сервер, отвечающий за:
   1) Регистрация и аутентификация пользователей;
   2) CRUD-операции над задачами;
   3) Отправка задач в очередь на выполнение.
  
2. `Task-Scheduler` - планировщик, запускающийся раз в N мс (значение N указывается в конфигурационном файле `config.json` и отправляющий задачи с отложенным запуском в очередь.

3. `Task-Worker` - компонент, получающий задачи из очереди и выполняющий их. Система предусматривает наличие несколько воркеров, работающих параллельно (их число указывается в переменной окружения `NUMOFWORKERS` в файле `.env`).

## Установка и запуск

1. Клонируйте репозиторий
   ```bash
   git clone https://github.com/imightbuyaboat/TaskFlow
   cd TaskFlow
   ```

2. Создайте файл окружения
   ```bash
   nano deployments/.env
   ```

   со следующим содержимым:
   ```.env
   POSTGRES_USER=your_user
   POSTGRES_PASSWORD=your_password
   POSTGRES_PORT=your_port
   POSTGRES_DB=your_database
    
   AMQP_USER=your_user
   AMQP_PASSWORD=your_password
   AMQP_PORT=your_port
   
   MAIL_HOST=your_mail_host
   MAIL_PORT=your_mail_port
   MAIL_USERNAME=your_username
   MAIL_PASSWORD=your_password
    
   SECRET_KEY=your_secret_key
    
   NUMOFWORKERS=3
    
   HOST_FILE_PATH=your_host_file_path
   BASE_FILE_PATH=your_base_file_path
   ```

3. Запустите сервис командой:
   ```bash
   docker compose -f deployments/docker-compose.yml --env-file deployments/.env up --build -d
   ```

## Типы (`Type`) и полезная нагрузка (`Payload`) задач, поддерживаемых системой

В список поддерживаемых системой задач входят:
1. Отправка писем:
   ```json
   "type": "send_email",
   "payload": {
            "to": "recipient's email address",
            "subject": "subject",
            "body": "body",
            "attached_files": ["your files"]
   }
   ```

   Поле `to`, а также одно из полей `subject`, `body`, `attached_files` являются обязательными.
  
2. Обработка изображений:
   ```json
   "type": "process_image",
   "payload": {
            "path": "file_name",
            "blur": 1,
            "sharpen": 1,
            "gamma": 1,
            "contrast": 1,
            "brightness": 1,
            "saturation": 1,
            "grayscale": true,
            "invert": true
   }
   ```

   Поле `path` является обязательным. Поля `blur`, `sharpen`, `gamma` могут принимать только положительные значения. Поля `contrast`, `brightness`, `saturation` могут принимать значения из диапазона [-100; 100].
   
3. Скачивание файлов по url:
   ```json
   "type": "download_files",
   "payload": {
            "urls": ["your urls"]
   }
   ```

## API-примеры (curl)

1. Регистрация
   ```bash
   curl -X POST http://localhost:8080/api/register \
   -H "Content-Type: application/json" \
   -d '{"email": "example@example.com", "password": "example"}'
   ```

2. Вход
   ```bash
   curl -X POST http://localhost:8080/api/login \
   -H "Content-Type: application/json" \
   -d '{"email": "example@example.com", "password": "example"}'
   ```
   
   В ответе на этот запрос будет возвращен JWT токен, который необходимо указывать в заголовке `Authorization` для всех последующих запросов.

3. Получение всех задач пользователя
   ```bash
   curl -X GET http://localhost:8080/api/tasks \
   -H "Authorization: your_token"
   ```

4. Получение задачи по ее `id`
   ```bash
   curl -X GET http://localhost:8080/api/tasks/task_id \
   -H "Authorization: your_token"
   ```

   Где `task_id` - это id задачи, возвращаемый после ее создания.

5. Создание задачи
   ```bash
   curl -X POST http://localhost:8080/api/tasks \
   -H "Authorization: your_token" \
   -H "Content-Type: application/json" \
   -d '{
     "type": "download_files",
     "payload": {
       "urls": ["https://go.dev/"]
     },
     "max_retries": 3,
     "run_at": "2025-06-03T12:50:50Z"
   }'
   ```

   Где `max_retries` - положительное число, указывающее на количество повторов выполнения задачи при ее неудачном выполнении (по умолчанию равно 3), `run_at` - время начала выполнения задачи (задачи с отложенным выполнением имеют статус `postponed`). Оба этих параметра являются необязательными при создании задачи.
