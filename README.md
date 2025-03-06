# Calculator

Финальный проект второго спринта годового курса Go от Я.Лицея

## Установка

 - Для установки нужно выбрать директорию проекта:
```bash
cd <your_dir>
```
 - Потом необходимо выполнить эту команду:
```bash
git clone https://github.com/xKARASb/Calculator
```
 - В выбранной папке появится папка ```Calculator``` c проектом.

## Использование

### Конфигурация
#### Переменные среды
Сначала необходимо открыть файл ```./config/.env``` и установить параметры:

 - **TIME_ADDITION_MS** - время вычисления сложения(в миллисекундах);

 - **TIME_SUBTRACTION_MS** - время вычисления вычитания;

 - **TIME_MULTIPLICATIONS_MS** - время вычисления умножения;

 - **TIME_DIVISIONS_MS** - время вычисления деления;

 - **COMPUTING_POWER** - максмальное количество *worker*'ов, которые параллельно выполняют арифметические действия.

#### Другие параметры

Потом необходимо открыть файл ```config.json``` в той же папке и установить следущие параметры(**true** - включено, **false** - выключено):

 - ```web``` - веб-интерфейс(подробнее в **Веб-интерфейс**)

По умолчанию выключено.

### Запуск
 - Для запуска API необходимо выбрать директорию проекта:
```
cd <путь к папке Calculator>
```
 - Далее надо запустить файл ```./cmd/main.go```:
```
go run ./cmd/main.go
```

## Использование API
#### Добавление вычисления арифметического выражения

##### Curl
```
curl --location 'localhost/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": <выражение>
}'
```
##### Коды ответа: 
 - 201 - выражение принято для вычисления
 - 422 - невалидные данные
 - 500 - что-то пошло не так

##### Тело ответа

```
{
    "id": <уникальный идентификатор выражения> // его ID
}
```
#### Все выражения
##### Curl
```
curl --location 'localhost/api/v1/expressions'
```
##### Тело ответа:

```
{
    "expressions": [
        {
            "id": 1,
            "status": "OK",
            "result": 3>
        },
        {
            "id": 1,
            "status": "Wait",
            "result": 0
        }
    ]
}
```
##### Коды ответа:
 - 200 - успешно получен список выражений
 - 500 - что-то пошло не так

#### Получение выражения по его id

Получение выражения по его идентификатору.
##### Curl

```
curl --location 'localhost/api/v1/expressions/<id>'
```

##### Тело ответа:

```
{
    "expression":
        {
            "id": <идентификатор выражения>,
            "status": <статус вычисления выражения>,
            "result": <результат выражения>
        }
}
```

##### Коды ответа:
 - 200 - успешно получено выражение
 - 404 - нет такого выражения
 - 500 - что-то пошло не так

## Примеры работы с API
#### Делаем запрос для вычисление выражения

##### Curl
```
curl --location 'localhost/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2/2"
}'
```

##### Ответ
Статус 201(успешно создано);
```
{
    "id": 12345
}
```

#### Конкретное выражение
##### Curl
```
curl --location 'localhost/api/v1/expressions/12345'
```

##### Ответ
Статус 200(успешно получено);
```
{
    "expression":
        {
            "id": 12345,
            "status": "OK",
            "result": 321
        }
}
```

#### Все выражения
##### Curl
```
curl --location 'localhost/api/v1/expressions'
```

##### Ответ
Статус 200(успешно получены);
```
{
    "expressions": [
        {
            "id": 12345,
            "status": "OK",
            "result": 321
        },
    ]
}
```

### Пример с ошибкой в запросе

#### Делаем **неправильный** запрос на вычисление выражения

##### Curl

```
curl --location 'localhost/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "adfsfds": "2+2/2"
}'
```

##### Ответ
Статус 422(**неправильный** запрос);


### Пример с ошибкой в запросе

#### Делаем **правильный** запрос на вычисление выражения

##### Curl
```
curl --location 'localhost/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2/2"
}'
```
##### Ответ
Статус 201(успешно создано);
```
{
    "id": 12345 // пример
}
```
#### Далее получаем наше выражение(**неправильный** ID)
##### Curl
```
curl --location 'localhost/api/v1/expressions/45362'
```

##### Ответ
Статус 404(не найдено);


### Пример с ошибкой в запросе

#### Делаем запрос с **некорректным** URL на вычисление выражения

##### Curl
```
curl --location 'localhost/api/v1/abc' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "121+2"
}'
```

##### Ответ
Статус 404(**NOT FOUND**);


### Веб-интерфейс

Вот ссылки на веб-страницы:

 - [Главная страница](http://localhost:8080/api/v1/web)

 - [Вычисление выражения](http://localhost:8080/api/v1/web/calculate)

 - [Просмотр выражений](http://localhost:8080/api/v1/web/expressions)

****ВАЖНО:**** По умолчанию веб-интерфейс выключен. Чтобы его включить, нужно изменить параметр *Веб интерфейс* в **Конфигурация/Другие Параметры**.