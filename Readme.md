# WB Tech L0
Демонстрационный сервис с простейшим интерфейсом, отображающий данные о заказе. Подписывается на топик в kafka, принимает сообщения, сохраняет данные в кэш и бд. По запросу от http сервера выдает данные из кэша, если они там есть, в противном случае идет в бд.
