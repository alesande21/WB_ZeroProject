package entity

import (
	"encoding/json"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CustomerId Идентификатор покупателя
type CustomerId = string

// Delivery Информация о доставке
type Delivery struct {
	// Address Адрес доставки
	Address DeliveryAddress `json:"address"`

	// City Город доствки
	City DeliveryCity `json:"city"`

	// Email Электронный адрес адресата доставки
	Email DeliveryEmail `json:"email"`

	// Id Уникальный идентификатор заказа
	Id OrderId `json:"id"`

	// Name Полное навзание доставки
	Name DeliveryName `json:"name"`

	// Phone Номер телофона
	Phone DeliveryPhone `json:"phone"`

	// Region Регион доставки
	Region DeliveryRegion `json:"region"`

	// Zip Почтовый индекс
	Zip DeliveryZip `json:"zip"`
}

// DeliveryAddress Адрес доставки
type DeliveryAddress = string

// DeliveryCity Город доствки
type DeliveryCity = string

// DeliveryEmail Электронный адрес адресата доставки
type DeliveryEmail = openapi_types.Email

// DeliveryName Полное навзание доставки
type DeliveryName = string

// DeliveryPhone Номер телофона
type DeliveryPhone = string

// DeliveryRegion Регион доставки
type DeliveryRegion = string

// DeliveryService Название случбы доставки
type DeliveryService = string

// DeliveryZip Почтовый индекс
type DeliveryZip = string

// Entry Код тороговой площадки или компании
type Entry = string

// ErrorResponse Используется для возвращения ошибки пользователю
type ErrorResponse struct {
	// Reason Описание ошибки в свободной форме
	Reason string `json:"reason"`
}

// Item Информация о товаре
type Item struct {
	// Brand Бренд товара.
	Brand ItemBrand `json:"brand"`

	// ChrtId Уникальный номер товара
	ChrtId ItemChrtId `json:"chrt_id"`

	// Name Название товара
	Name ItemName `json:"name"`

	// NmId Внутренний идентификатор товара на торговой площадке или в системе продаж
	NmId ItemNmId `json:"nm_id"`

	// Price Цена товара до скидки
	Price ItemPrice `json:"price"`

	// Rid Уникальный идентификатор записи товара в системе
	Rid ItemRid `json:"rid"`

	// Sale Размер скидки на товар в процентах
	Sale ItemSale `json:"sale"`

	// Size Размер товара
	Size ItemSize `json:"size"`

	// Status Статус товара или заказа
	Status ItemStatus `json:"status"`

	// TotalPrice Итоговая цена с учетом скидки
	TotalPrice ItemTotalPrice `json:"total_price"`

	// TrackNumber Номер отслеживания товара
	TrackNumber TrackNumber `json:"track_number"`
}

// ItemBrand Бренд товара.
type ItemBrand = string

// ItemChrtId Уникальный номер товара
type ItemChrtId = int64

// ItemName Название товара
type ItemName = string

// ItemNmId Внутренний идентификатор товара на торговой площадке или в системе продаж
type ItemNmId = int64

// ItemPrice Цена товара до скидки
type ItemPrice = float64

// ItemRid Уникальный идентификатор записи товара в системе
type ItemRid = string

// ItemSale Размер скидки на товар в процентах
type ItemSale = float64

// ItemSize Размер товара
type ItemSize = string

// ItemStatus Статус товара или заказа
type ItemStatus = int32

// ItemTotalPrice Итоговая цена с учетом скидки
type ItemTotalPrice = float64

// Items Список товаров
type Items = []Item

// Order Детали заказа
type Order struct {
	// CustomerId Идентификатор покупателя
	CustomerId CustomerId `json:"customer_id"`

	// DateCreated Дата и время создания заказа
	DateCreated OrderDateCreated `json:"date_created"`

	// Delivery Информация о доставке
	Delivery Delivery `json:"delivery"`

	// DeliveryService Название случбы доставки
	DeliveryService DeliveryService `json:"delivery_service"`

	// Entry Код тороговой площадки или компании
	Entry Entry `json:"entry"`

	// InternalSignature Внутренняя подпись для проверки или идентификации заказа
	InternalSignature OrderInternalSignature `json:"internal_signature"`

	// Items Список товаров
	Items Items `json:"items"`

	// Locale Язык или региональные настройки
	Locale OrderLocale `json:"locale"`

	// OofShard Дополнительное поле для шардирования
	OofShard OrderOofShard `json:"oof_shard"`

	// OrderUid Уникальный идентификатор заказа
	OrderUid OrderId `json:"order_uid"`

	// Payment Информация о оплате заказа
	Payment Payment `json:"payment"`

	// Shardkey Ключ шардирования для распределения данных по различным серверам или базам данных.
	Shardkey ShardKey `json:"shardkey"`

	// SmId Идентификатор партнера или магазина.
	SmId SmId `json:"sm_id"`

	// TrackNumber Номер отслеживания товара
	TrackNumber TrackNumber `json:"track_number"`
}

// OrderDateCreated Дата и время создания заказа
type OrderDateCreated = string

// OrderId Уникальный идентификатор заказа
type OrderId = string

// OrderInternalSignature Внутренняя подпись для проверки или идентификации заказа
type OrderInternalSignature = string

// OrderLocale Язык или региональные настройки
type OrderLocale = string

// OrderOofShard Дополнительное поле для шардирования
type OrderOofShard = string

// Payment Информация о оплате заказа
type Payment struct {
	// Amount Сумма транзакции
	Amount PaymentAmount `json:"amount"`

	// Bank Банк транзакции
	Bank PaymentBank `json:"bank"`

	// Currency Валюта транзакции
	Currency PaymentCurrency `json:"currency"`

	// CustomFee Дополнительные расходы на доставку
	CustomFee PaymentCustomFee `json:"custom_fee"`

	// DeliveryCost Стоимость доставки
	DeliveryCost PaymentDeliveryCost `json:"delivery_cost"`

	// GoodsTotal Общая стоимость товаров без учета доставки
	GoodsTotal PaymentGoodsTotal `json:"goods_total"`

	// PaymentDt Время транзакции в формате Unix-времени
	PaymentDt PaymentDt `json:"payment_dt"`

	// Provider Поставщик
	Provider PaymentProvider `json:"provider"`

	// RequestId Номер транзакции
	RequestId PaymentRequestId `json:"request_id"`

	// Transaction Уникальный идентификатор заказа
	Transaction OrderId `json:"transaction"`
}

// PaymentAmount Сумма транзакции
type PaymentAmount = float64

// PaymentBank Банк транзакции
type PaymentBank = string

// PaymentCurrency Валюта транзакции
type PaymentCurrency = string

// PaymentCustomFee Дополнительные расходы на доставку
type PaymentCustomFee = float64

// PaymentDeliveryCost Стоимость доставки
type PaymentDeliveryCost = float64

// PaymentDt Время транзакции в формате Unix-времени
type PaymentDt = string

// PaymentGoodsTotal Общая стоимость товаров без учета доставки
type PaymentGoodsTotal = float64

// PaymentProvider Поставщик
type PaymentProvider = string

// PaymentRequestId Номер транзакции
type PaymentRequestId = string

// ShardKey Ключ шардирования для распределения данных по различным серверам или базам данных.
type ShardKey = string

// SmId Идентификатор партнера или магазина.
type SmId = int64

// TrackNumber Номер отслеживания товара
type TrackNumber = string

// PaginationLimit defines model for paginationLimit.
type PaginationLimit = int32

// PaginationOffset defines model for paginationOffset.
type PaginationOffset = int32

// GetOrdersParams defines parameters for GetOrders.
type GetOrdersParams struct {
	// Limit Максимальное число возвращаемых объектов. Используется для запросов с пагинацией.
	//
	// Сервер должен возвращать максимальное допустимое число объектов.
	Limit *PaginationLimit `form:"limit,omitempty" json:"limit,omitempty"`

	// Offset Какое количество объектов должно быть пропущено с начала. Используется для запросов с пагинацией.
	Offset *PaginationOffset `form:"offset,omitempty" json:"offset,omitempty"`
}

// CreateOrderJSONBody defines parameters for CreateOrder.
type CreateOrderJSONBody struct {
	union json.RawMessage
}

// CreateOrderJSONBody0 defines parameters for CreateOrder.
type CreateOrderJSONBody0 = []Order

// CreateOrderJSONRequestBody defines body for CreateOrder for application/json ContentType.
type CreateOrderJSONRequestBody CreateOrderJSONBody
