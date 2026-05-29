package constants

const (
	StatusNormal  = 0
	StatusDeleted = 1
)

// 司机工作状态
const (
	WorkStatusOffline   int8 = 0 // 离线
	WorkStatusOnline    int8 = 1 // 在线
	WorkStatusListening int8 = 2 // 听单中
)

// 账号状态
const (
	AccountStatusNormal    int8 = 1 // 正常
	AccountStatusFrozen    int8 = 2 // 冻结
	AccountStatusCancelled int8 = 3 // 已注销
)

// 认证状态（实名/车辆/资质通用）
const (
	AuthStatusNotSubmitted int8 = 0 // 未提交
	AuthStatusVerifying    int8 = 1 // 审核中
	AuthStatusApproved     int8 = 2 // 已通过
	AuthStatusRejected     int8 = 3 // 审核失败
)

// 订单状态
const (
	OrderStatusPending    int8 = 0 // 待派单
	OrderStatusDispatched int8 = 1 // 已派单
	OrderStatusAccepted   int8 = 2 // 司机已接单
	OrderStatusArrived    int8 = 3 // 司机已到达
	OrderStatusInTrip     int8 = 4 // 行程中
	OrderStatusCompleted  int8 = 5 // 已完成
	OrderStatusCancelled  int8 = 6 // 已取消
	OrderStatusPool       int8 = 7 // 抢单池中
)

// 派单拒绝原因
const (
	DispatchNotListening    = 3001 // 司机未在听单
	DispatchTooFar          = 3002 // 距离超过半径
	DispatchVehicleMismatch = 3003 // 车型不匹配
	DispatchHasOngoingOrder = 3004 // 有进行中订单
	DispatchDailyLimit      = 3005 // 达到接单上限
	DispatchLowScore        = 3006 // 服务分不足
)

// 上线校验错误码
const (
	ErrCodeOnlineNotLoggedIn   = 2001
	ErrCodeRealnameNotApproved = 2002
	ErrCodeVehicleNotApproved  = 2003
	ErrCodeAccountAbnormal     = 2005
	ErrCodeAlreadyOnline       = 2006
	ErrCodeHasOngoingOrder     = 2007
)

// 开始行程错误码
const (
	ErrCodePhoneMismatch = 2010 // 手机号后四位不匹配
)

// 拒绝接单原因码
const (
	RejectReasonTooFar      = 4001 // 距离太远
	RejectReasonDirection   = 4002 // 方向不符
	RejectReasonAlreadyBusy = 4003 // 已有订单
	RejectReasonOther       = 4004 // 其他原因
)

// 抢单池相关
const (
	PoolOrderMaxSize     = 30  // 抢单池最大订单数
	PoolOrderTimeoutSec  = 1800 // 抢单超时时间(秒) = 30分钟
	PoolPageSizeDefault  = 20  // 默认每页数量
	PoolPageSizeMax      = 30  // 最大每页数量
)

// 抢单错误码
const (
	GrabErrOrderNotFound  = 5001 // 订单不存在
	GrabErrNotInPool      = 5002 // 订单不在抢单池中
	GrabErrGrabbed        = 5003 // 订单已被抢走
	GrabErrTimeout        = 5004 // 抢单超时
	GrabErrSelf           = 5005 // 不能抢自己的订单
	GrabErrDailyLimit     = 5006 // 达到当日接单上限
	GrabErrCityMismatch   = 5007 // 城市不匹配
)

// 进池原因码（记录订单为何进入抢单池）
const (
	PoolReasonNoDriver        = 1 // 附近无可用司机
	PoolReasonAllRejected     = 2 // 所有司机拒绝
	PoolReasonRejectExceed    = 3 // 拒绝次数超过阈值
	PoolReasonDispatchTimeout = 4 // 派单超时未响应
	PoolReasonExpandedFail    = 5 // 扩大半径重试仍无司机
)
