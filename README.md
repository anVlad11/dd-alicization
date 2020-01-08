# dd-alicization

Go-модуль для работы с Яндекс.Станцией

Установка:

``
go get github.com/anvlad11/dd-alicization
``

Модуль содержит пакет glagol
* glagol.APIClient - работа с API Яндекса для получения списка устройств пользователя;
* glagol.Conversation - работа с Яндекс.Станцией через веб-сокеты;
* glagol.HttpGateway - HTTP-gateway для проброса запросов из внешних сервисов к Яндекс.Станции;
