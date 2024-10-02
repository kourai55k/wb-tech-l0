package models

// Order структура для заказа
type Order struct {
	OrderUID          string   `json:"order_uid" validate:"required"`                                       // UUID4
	TrackNumber       string   `json:"track_number" validate:"required"`                                    // Обязательное поле
	Entry             string   `json:"entry" validate:"required"`                                           // Обязательное поле
	Delivery          Delivery `json:"delivery" validate:"required,dive"`                                   // Обязательное поле с проверкой по элементам
	Payment           Payment  `json:"payment" validate:"required,dive"`                                    // Обязательное поле с проверкой по элементам
	Items             []Item   `json:"items" validate:"required,dive,min=1"`                                // Обязательное поле, минимум 1 элемент
	Locale            string   `json:"locale" validate:"required"`                                          // Обязательное поле
	InternalSignature string   `json:"internal_signature"`                                                  // Необязательное поле
	CustomerID        string   `json:"customer_id" validate:"required"`                                     // Обязательное поле
	DeliveryService   string   `json:"delivery_service" validate:"required"`                                // Обязательное поле
	Shardkey          string   `json:"shardkey" validate:"required"`                                        // Обязательное поле
	SmID              int      `json:"sm_id" validate:"required,min=1"`                                     // Обязательное поле, минимум 1
	DateCreated       string   `json:"date_created" validate:"required,datetime=2006-01-02T15:04:05Z07:00"` // Дата в определенном формате
	OofShard          string   `json:"oof_shard" validate:"required"`                                       // Обязательное поле
}

// Delivery структура для доставки
type Delivery struct {
	Name    string `json:"name" validate:"required"`              // Обязательное поле
	Phone   string `json:"phone" validate:"required,e164"`        // Обязательное поле, формат E.164
	Zip     string `json:"zip" validate:"required,numeric,len=7"` // Обязательное поле, 7 цифр
	City    string `json:"city" validate:"required"`              // Обязательное поле
	Address string `json:"address" validate:"required"`           // Обязательное поле
	Region  string `json:"region" validate:"required"`            // Обязательное поле
	Email   string `json:"email" validate:"required,email"`       // Обязательное поле, формат email
}

// Payment структура для платежа
type Payment struct {
	Transaction  string `json:"transaction" validate:"required"`  // Обязательное поле
	RequestID    string `json:"request_id"`                       // Необязательное поле
	Currency     string `json:"currency" validate:"required"`     // Обязательное поле
	Provider     string `json:"provider" validate:"required"`     // Обязательное поле
	Amount       int    `json:"amount" validate:"required,gte=0"` // Обязательное поле, больше или равно 0
	PaymentDT    int64  `json:"payment_dt" validate:"required"`   // Обязательное поле
	Bank         string `json:"bank" validate:"required"`         // Обязательное поле
	DeliveryCost int    `json:"delivery_cost" validate:"gte=0"`   // Необязательное поле, больше или равно 0
	GoodsTotal   int    `json:"goods_total" validate:"gte=0"`     // Необязательное поле, больше или равно 0
	CustomFee    int    `json:"custom_fee" validate:"gte=0"`      // Необязательное поле, больше или равно 0
}

// Item структура для товара
type Item struct {
	ChrtID      int    `json:"chrt_id" validate:"required"`           // Обязательное поле
	TrackNumber string `json:"track_number" validate:"required"`      // Обязательное поле
	Price       int    `json:"price" validate:"required,gte=0"`       // Обязательное поле, больше или равно 0
	Rid         string `json:"rid" validate:"required"`               // Обязательное поле
	Name        string `json:"name" validate:"required"`              // Обязательное поле
	Sale        int    `json:"sale" validate:"gte=0"`                 // Необязательное поле, больше или равно 0
	Size        string `json:"size" validate:"required"`              // Обязательное поле
	TotalPrice  int    `json:"total_price" validate:"required,gte=0"` // Обязательное поле, больше или равно 0
	NmID        int    `json:"nm_id" validate:"required"`             // Обязательное поле
	Brand       string `json:"brand" validate:"required"`             // Обязательное поле
	Status      int    `json:"status" validate:"required"`            // Обязательное поле
}
