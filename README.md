# Лабораторная работа #2

![GitHub Classroom Workflow](../../workflows/GitHub%20Classroom%20Workflow/badge.svg?branch=master)

## Microservices

### Формулировка

В рамках второй лабораторной работы _по вариантам_ требуется реализовать систему, состоящую из нескольких
взаимодействующих друг с другом сервисов.

### Требования

1. Каждый сервис имеет свое собственное хранилище, если оно ему нужно. Для учебных целей можно использовать один
   instance базы данных, но каждый сервис работает _только_ со своей логической базой. Запросы между базами _запрещены_.
2. Для межсервисного взаимодействия использовать HTTP (придерживаться RESTful). Допускается использовать и другие
   протоколы, например grpc, но это требуется согласовать с преподавателем.
3. Выделить **Gateway Service** как единую точку входа и межсервисной коммуникации. Горизонтальные запросы между
   сервисами делать _нельзя_.
4. На каждом сервисе сделать специальный endpoint `GET /manage/health`, отдающий 200 ОК, он будет использоваться для
   проверки доступности сервиса (в [Github Actions](.github/workflows/classroom.yml) в скрипте проверки готовности всех
   сервисов [wait-script.sh](scripts/wait-script.sh).
   ```shell
   "$path"/wait-for.sh -t 120 "http://localhost:$port/manage/health" -- echo "Host localhost:$port is active"
   ```
6. Код хранить на Github, для сборки использовать Github Actions.
7. Gateway Service должен запускаться на порту 8080, остальные сервисы запускать на портах 8050, 8060, 8070.
8. Каждый сервис должен быть завернут в docker.
9. В [docker-compose.yml](docker-compose.yml) прописать сборку и запуск docker контейнеров.
10. В [classroom.yml](.github/workflows/classroom.yml) дописать шаги на сборку и прогон unit-тестов.
11. Для автоматических прогонов тестов в файле [autograding.json](.github/classroom/autograding.json)
    и [classroom.yml](.github/workflows/classroom.yml) заменить `<variant>` на ваш вариант.

### Пояснения

1. Для разработки можно использовать Postgres в docker, для этого нужно запустить docker compose up -d, поднимется
   контейнер с Postgres 13, и будут созданы соответствующие вашему варианту (описанные в
   файлах [schema-$VARIANT](postgres/scripts)) базы данных и пользователь `program`:`test`.
2. Для создания базы нужно прописать в [20-create-schemas.sh](postgres/20-create-databases.sh) свой вариант задания в
3. Docker Compose позволяет выполнять сборку образа, для этого нужно прописать
   блок [`build`](https://docs.docker.com/compose/compose-file/build/).
4. Горизонтальную коммуникацию между сервисами делать нельзя.
5. Интеграционные тесты можно проверить локально, для этого нужно импортировать в Postman
   коллекцию `<variant>/postman/collection.json`) и `<variant>/postman/environment.json`.

![Services](images/services.png)

Предположим, у нас сервисы `UserService`, `OrderService`, `WarehouseService` и `Gateway`:

* На `Gateway` от пользователя `Alex` приходит запрос `Купить товар с productName: 'Lego Technic 42129`.
* `Gateway` -> `UserService` проверяем что пользователь существует и получаем `userUid` пользователя по `login: Alex`.
* `Gateway` -> `WarehouseService` получаем `itemUid` товара по `productName` и резервируем его для заказа.
* `Gateway` -> `OrderService` с `userUid` и `itemUid` и создаем заказ с `orderUid`.
* `Gateway` -> `WarehouseService` с `orderUid` и переводим товар `itemUid` из статуса `Зарезервировано` в
  статус `Заказан` и прописываем ссылку на `orderUid`.

### Прием задания

1. При получении задания у вас создается fork этого репозитория для вашего пользователя.
2. После того как все тесты успешно завершатся, в Github Classroom на Dashboard будет отмечено успешное выполнение
   тестов.

### Варианты заданий

Варианты заданий берутся исходя из формулы:
(номер в [списке группы](https://docs.google.com/spreadsheets/d/1BT5iLgERiWUPPn4gtOQk4KfHjVOTQbUS7ragAJrl6-Q)-1) % 4)+1.

1. [Flight Booking System](v1/README.md)
1. [Hotels Booking System](v2/README.md)
1. [Car Rental System](v3/README.md)
1. [Library System](v4/README.md)


# Лабораторная работа #3

![GitHub Classroom Workflow](../../workflows/GitHub%20Classroom%20Workflow/badge.svg?branch=master)

## Fault Tolerance

### Формулировка

На базе [Лабораторной работы #2](https://github.com/bmstu-rsoi/lab2-template) реализовать механизмы, увеличивающие
отказоустойчивость системы.

### Требования

1. На Gateway Service для _всех операций_ чтения реализовать паттерн Circuit Breaker. Накапливать статистику в памяти, и
   если система не ответила N раз, то в N + 1 раз вместо запроса сразу отдавать fallback. Через небольшой timeout
   выполнить запрос к реальной системе, чтобы проверить ее состояние.
2. На каждом сервисе сделать специальный endpoint `GET /manage/health`, отдающий 200 ОК, он будет использоваться для
   проверки доступности сервиса (в [Github Actions](.github/workflows/classroom.yml) в скрипте проверки готовности всех
   сервисов [wait-script.sh](scripts/wait-script.sh) и в тестах [test-script.sh](scripts/test-script.sh)).
   ```shell
   "$path"/wait-for.sh -t 120 "http://localhost:$port/manage/health" -- echo "Host localhost:$port is active"
   ```
4. В случае недоступности данных из некритичного источника (не основного), возвращается fallback-ответ. В зависимости от
   ситуации, это может быть:
   * пустой объект или массив;
   * объект, с заполненным полем (`uid` или подобным), по которому идет связь с другой системой;
   * default строка (если при этом не меняется тип переменной).
5. В задании описаны две операции, изменяющие состояния нескольких систем. В случае недоступности одной из систем,
   участвующих в этой операции, выполнить:
   1. откат всей операции;
   2. возвращать пользователю ответ об успешном завершении операции, а на Gateway Service поставить этот запрос в
      очередь для повторного выполнения.
6. Для автоматических прогонов тестов в файле [autograding.json](.github/classroom/autograding.json)
   и [classroom.yml](.github/workflows/classroom.yml) заменить `<variant>` на ваш вариант.
7. В [docker-compose.yml](docker-compose.yml) прописать сборку и запуск docker контейнеров.
8. Код хранить на Github, для сборки использовать Github Actions.
9. Каждый сервис должен быть завернут в docker.
10. В classroom.yml дописать шаги на сборку, прогон unit-тестов.

### Пояснения

1. Для локальной разработки можно использовать Postgres в docker.
2. Схема взаимодействия сервисов остается как в [Лабораторной работы #2](https://github.com/bmstu-rsoi/lab2-template).
3. Для реализации очереди можно использовать language native реализацию (например, BlockingQueue для Java), либо
   какую-то готовую реализацию типа RabbitMQ, Redis, ZeroMQ и т.п. Крайне нежелательно использовать реляционную базу
   данных как средство эмуляции очереди.
4. Можно использовать внешнюю очередь или запускать ее в docker.
5. Для проверки отказоустойчивости используется остановка и запуск контейнеров docker, это делает
   скрипт [test-script.sh](scripts/test-script.sh). Скрипт нужно запускать из корня проекта, т.к. он обращается к папке
   postman по вариантам.
   ```shell
   # запуск тестового сценария:
   # * <variant> – номер варианта (v1 | v2 | v3 | v4 )
   # * <service> – имя сервиса в Docker Compose
   # * <port>    – порт, на котором запущен сервис
   $ scripts/test-script.sh <variant> <service> <port>
   ```


# Лабораторная работа #4

![GitHub Classroom Workflow](../../workflows/GitHub%20Classroom%20Workflow/badge.svg?branch=master)

## Deploy to Cloud

### Формулировка

На базе [Лабораторной работы #2](https://github.com/bmstu-rsoi/lab2-template) выполнить деплой приложения в managed
кластер k8s.

### Требования

1. Скопировать исходный код из ЛР #2 в проект.
2. Развернуть руками свой Managed Kubernetes Cluster, настроить Ingress Controller (для публикации сервисов наружу можно
   использовать _только_ Ingress).
4. Собрать и опубликовать образы docker в [Docker Registry](https://hub.docker.com/).
5. Описать манифесты для деплоя в виде [helm charts](https://helm.sh/docs/topics/charts/), они должен быть универсальным
   для всех сервисов и отличаться лишь набором параметров запуска.
6. В кластере k8s можно использовать один физический instance базы, но каждый сервис должен работать только со своей
   виртуальной базой данных. Задеплоить базу в кластер можно руками, либо использовать уже готовый helm chart.
7. Код хранить на Github, для сборки использовать Github Actions.
8. Для автоматических прогонов тестов в файле [autograding.json](.github/classroom/autograding.json)
   и [classroom.yml](.github/workflows/classroom.yml) заменить `<variant>` на ваш вариант.
9. В [classroom.yml](.github/workflows/classroom.yml) дописать шаги:
   1. сборка приложения;
   2. сборка и публикация образа docker (можно использовать `docker compose build`, `docker compose push`);
   3. деплой каждого сервиса в кластер k8s.

### Пояснения

Т.к. развертывание полноценного кластера на виртуальным машинах очень сложный процесс, можно использовать Managed
Kubernetes Cluster, т.е. готовый кластер k8s, предоставляемый сторонней платформой, например:

* [Digital Ocean](https://www.digitalocean.com/products/kubernetes/)
* [Yandex Cloud](https://cloud.yandex.ru/services/managed-kubernetes)
* [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine)
* [AWS](https://aws.amazon.com/ru/eks/)

Платформ, которые предоставляют Kubernetes as a Service большое количество, вы можете сами исследовать рынок и выбрать
другого провайдера услуг. Большинство провайдеров имеют бесплатный триальный период или денежный грант.

Для создания кластера достаточно 2-3 worker ноды 2Gb, 1CPU.


# Лабораторная работа #5

![GitHub Classroom Workflow](../../workflows/GitHub%20Classroom%20Workflow/badge.svg?branch=master)

## OAuth2 Authorization

### Формулировка

На базе [Лабораторной работы #4](https://github.com/bmstu-rsoi/lab2-template) реализовать OAuth2 token-based
авторизацию.

* Для авторизации использовать OpenID Connect, в роли Identity Provider использовать стороннее решение.
* На Identity Provider настроить
  использование [Resource Owner Password flow](https://auth0.com/docs/authorization/flows/resource-owner-password-flow)
  (в одном запросе передается `clientId`, `clientSecret`, `username`, `password`).
* Все методы `/api/**` (кроме `/api/v1/authorize` и `/api/v1/callback`) на всех сервисах закрыть token-based
  авторизацией.
* В качестве токена использовать [JWT](https://jwt.io/introduction), для валидации токена
  использовать [JWKs](https://auth0.com/docs/security/tokens/json-web-tokens/json-web-key-sets), _запрос к Identity
  Provider делать не нужно_.
* JWT токен пробрасывать между сервисами, при получении запроса валидацию токена так же реализовать через JWKs.
* Убрать заголовок `X-User-Name` и получать пользователя из JWT-токена.
* Если авторизация некорректная (отсутствие токена, ошибка валидации JWT токена, закончилось время жизни токена
  (поле `exp` в payload)), то отдавать 401 ошибку.
* В `scope` достаточно указывать `openid profile email`.

### Требования

1. Для автоматических прогонов тестов в файле [autograding.json](.github/classroom/autograding.json)
   и [classroom.yml](.github/workflows/classroom.yml) заменить `<variant>` на ваш вариант.
1. Код хранить на Github, для сборки использовать Github Actions.
1. Каждый сервис должен быть завернут в docker.
1. В classroom.yml дописать шаги на сборку, прогон unit-тестов.

### Пояснения

1. В роли Identity Provider можно использовать любое решение, вот несколько рабочих вариантов:
   1. [Okta](https://developer.okta.com/docs/guides/)
   2. [Auth0](https://auth0.com/developers)
2. Для получения metadata для OpenID Connect можно
   использовать [Well-Known URI](https://auth0.com/docs/security/tokens/json-web-tokens/locate-json-web-key-sets):
   `https://[base-server-url]/.well-known/openid-configuration`.
3. Из Well-Known metadata можно получить Issuer URI и JWKs URI.
4. Для реализации OAuth2 можно использовать сторонние библиотеки.


### Прием задания

1. При получении задания у вас создается fork этого репозитория для вашего пользователя.
2. После того как все тесты успешно завершатся, в Github Classroom на Dashboard будет отмечено успешное выполнение
   тестов.

### Варианты заданий

Распределение вариантов заданий аналогично [ЛР #2](https://github.com/bmstu-rsoi/lab2-template).

1. [Flight Booking System](v1/README.md)
2. [Hotels Booking System](v2/README.md)
3. [Car Rental System](v3/README.md)
4. [Library System](v4/README.md)