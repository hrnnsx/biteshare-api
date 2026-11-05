package model

type ProductStatus string

const (
	ProductAvailable   ProductStatus = "available"
	ProductSoldOut     ProductStatus = "sold_out"
	ProductUnavailable ProductStatus = "unavailable"
	ProductExpired     ProductStatus = "expired"
)

type OrderStatus string

const (
	OrderPending        OrderStatus = "pending"
	OrderWaitingPayment OrderStatus = "waiting_payment"
	OrderPaid           OrderStatus = "paid"
	OrderProcessing     OrderStatus = "processing"
	OrderShipped        OrderStatus = "shipped"
	OrderCompleted      OrderStatus = "completed"
	OrderCancelled      OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentPending  PaymentStatus = "pending"
	PaymentSuccess  PaymentStatus = "success"
	PaymentFailed   PaymentStatus = "failed"
	PaymentExpired  PaymentStatus = "expired"
	PaymentRefunded PaymentStatus = "refunded"
)

type DeliveryStatus string

const (
	DeliveryPending    DeliveryStatus = "pending"
	DeliveryPickedUp   DeliveryStatus = "picked_up"
	DeliveryDelivering DeliveryStatus = "delivering"
	DeliveryDelivered  DeliveryStatus = "delivered"
	DeliveryFailed     DeliveryStatus = "failed"
	DeliveryCancelled  DeliveryStatus = "cancelled"
)
