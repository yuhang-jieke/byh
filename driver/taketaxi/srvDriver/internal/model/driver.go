package model

import "time"

type Driver struct {
	ID     uint   `gorm:"primaryKey"`
	Name   string `gorm:"size:255"`
	Status int    `gorm:"default:0"`
}

func (Driver) TableName() string { return "" }

// Passenger 乘客基础信息表
type Passenger struct {
	PassengerId    int64   `gorm:"column:passenger_id;primaryKey;comment:乘客ID，分布式ID" json:"passenger_id"`
	Mobile         string  `gorm:"column:mobile;size:20;not null;comment:登录手机号" json:"mobile"`
	MobileEncrypt  string  `gorm:"column:mobile_encrypt;size:100;comment:手机号AES加密" json:"mobile_encrypt"`
	Nickname       string  `gorm:"column:nickname;size:64;comment:乘客昵称" json:"nickname"`
	Avatar         string  `gorm:"column:avatar;size:512;comment:头像URL" json:"avatar"`
	Gender         int8    `gorm:"column:gender;default:0;comment:性别: 0-未知 1-男 2-女" json:"gender"`
	Status         int8    `gorm:"column:status;not null;default:1;comment:账号状态: 1-正常 2-冻结 3-已注销" json:"status"`
	VerifyStatus   int8    `gorm:"column:verify_status;not null;default:0;comment:实名认证状态: 0-未认证 1-认证中 2-已认证" json:"verify_status"`
	Level          int8    `gorm:"column:level;not null;default:1;comment:会员等级: 1-普通 2-白银 3-黄金 4-铂金 5-钻石" json:"level"`
	TotalConsumed  float64 `gorm:"column:total_consumed;type:decimal(12,2);not null;default:0.00;comment:累计消费金额" json:"total_consumed"`
	OrderCount     int     `gorm:"column:order_count;not null;default:0;comment:累计订单数" json:"order_count"`
	RegisterSource string  `gorm:"column:register_source;size:32;comment:注册来源" json:"register_source"`
	CityId         int64   `gorm:"column:city_id;comment:常用城市ID" json:"city_id"`
	LastOrderAt    string  `gorm:"column:last_order_at;comment:最后下单时间" json:"last_order_at"`
	CreatedAt      string  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt      string  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
	DeletedAt      string  `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
}

func (Passenger) TableName() string {
	return "passenger"
}

// / Driver 司机基础信息表
type DriverS struct {
	DriverId        int64   `gorm:"column:driver_id;primaryKey;comment:司机ID，分布式ID" json:"driver_id"`
	Mobile          string  `gorm:"column:mobile;size:20;not null;comment:登录手机号" json:"mobile"`
	MobileEncrypt   string  `gorm:"column:mobile_encrypt;size:100;comment:手机号AES加密" json:"mobile_encrypt"`
	Nickname        string  `gorm:"column:nickname;size:64;comment:司机昵称" json:"nickname"`
	Avatar          string  `gorm:"column:avatar;size:512;comment:头像URL" json:"avatar"`
	Gender          int8    `gorm:"column:gender;default:0;comment:性别: 0-未知 1-男 2-女" json:"gender"`
	Status          int8    `gorm:"column:status;not null;default:1;comment:账号状态: 1-正常 2-冻结 3-已注销" json:"status"`
	VerifyStatus    int8    `gorm:"column:verify_status;not null;default:0;comment:认证状态: 0-未认证 1-认证中 2-已认证 3-认证失败" json:"verify_status"`
	ServiceScore    float64 `gorm:"column:service_score;type:decimal(3,1);not null;default:80.0;comment:服务评分" json:"service_score"`
	OrderCount      int     `gorm:"column:order_count;not null;default:0;comment:累计完成订单数" json:"order_count"`
	TotalIncome     float64 `gorm:"column:total_income;type:decimal(12,2);not null;default:0.00;comment:累计收入金额" json:"total_income"`
	RegisterSource  string  `gorm:"column:register_source;size:32;comment:注册来源" json:"register_source"`
	CityId          int64   `gorm:"column:city_id;comment:服务城市ID" json:"city_id"`
	LastOnlineAt    string  `gorm:"column:last_online_at;comment:最后在线时间" json:"last_online_at"`
	WorkStatus      int64   `gorm:"column:work_status;comment:工作状态" json:"work_status"`
	DailyOrderLimit int     `gorm:"column:daily_order_limit;comment:接单上限" json:"daily_order_limit"`
	CreatedAt       string  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt       string  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
	DeletedAt       string  `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
}

func (DriverS) TableName() string {
	return "drivers"
}

// ----------------------------
// 3. driver_realname 司机实名认证表
// ----------------------------
type DriverRealname struct {
	Id              int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId        int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	RealName        string    `gorm:"column:real_name;comment:真实姓名" json:"real_name"`
	RealNameEncrypt string    `gorm:"column:real_name_encrypt;comment:姓名AES加密" json:"real_name_encrypt"`
	IdCardNo        string    `gorm:"column:id_card_no;comment:身份证号" json:"id_card_no"`
	IdCardNoEncrypt string    `gorm:"column:id_card_no_encrypt;comment:身份证号AES加密" json:"id_card_no_encrypt"`
	IdCardFrontUrl  string    `gorm:"column:id_card_front_url;comment:身份证正面照片URL" json:"id_card_front_url"`
	IdCardBackUrl   string    `gorm:"column:id_card_back_url;comment:身份证反面照片URL" json:"id_card_back_url"`
	Gender          int8      `gorm:"column:gender;comment:性别: 1-男 2-女" json:"gender"`
	Birthday        time.Time `gorm:"column:birthday;comment:出生日期" json:"birthday"`
	Address         string    `gorm:"column:address;comment:身份证地址" json:"address"`
	Nation          string    `gorm:"column:nation;comment:民族" json:"nation"`
	ExpireDate      time.Time `gorm:"column:expire_date;comment:身份证有效期" json:"expire_date"`
	Status          int8      `gorm:"column:status;comment:认证状态: 0-未认证 1-认证中 2-已认证 3-认证失败" json:"status"`
	FailReason      string    `gorm:"column:fail_reason;comment:认证失败原因" json:"fail_reason"`
	VerifyTime      time.Time `gorm:"column:verify_time;comment:认证完成时间" json:"verify_time"`
	CreatedAt       time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverRealname) TableName() string {
	return "driver_realname"
}

// ----------------------------
// 4. driver_license 司机驾驶证认证表
// ----------------------------
type DriverLicense struct {
	Id             int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId       int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	LicenseNo      string    `gorm:"column:license_no;comment:驾驶证编号" json:"license_no"`
	LicenseType    string    `gorm:"column:license_type;comment:准驾车型: C1/C2/B1/B2/A1/A2" json:"license_type"`
	LicenseUrl     string    `gorm:"column:license_url;comment:驾驶证照片URL" json:"license_url"`
	FirstIssueDate time.Time `gorm:"column:first_issue_date;comment:初次领证日期" json:"first_issue_date"`
	IssueDate      time.Time `gorm:"column:issue_date;comment:当前证件发证日期" json:"issue_date"`
	ExpireDate     time.Time `gorm:"column:expire_date;comment:驾驶证有效期" json:"expire_date"`
	DrivingYears   int8      `gorm:"column:driving_years;comment:驾龄(年)" json:"driving_years"`
	Status         int8      `gorm:"column:status;comment:认证状态: 0-未认证 1-认证中 2-已认证 3-认证失败" json:"status"`
	FailReason     string    `gorm:"column:fail_reason;comment:认证失败原因" json:"fail_reason"`
	VerifyTime     time.Time `gorm:"column:verify_time;comment:认证完成时间" json:"verify_time"`
	CreatedAt      time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverLicense) TableName() string {
	return "driver_license"
}

// ----------------------------
// 5. driver_vehicle 司机车辆信息表
// ----------------------------
type DriverVehicle struct {
	Id                int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId          int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	PlateNo           string    `gorm:"column:plate_no;comment:车牌号" json:"plate_no"`
	PlateNoEncrypt    string    `gorm:"column:plate_no_encrypt;comment:车牌号AES加密" json:"plate_no_encrypt"`
	VehicleModel      string    `gorm:"column:vehicle_model;comment:车型名称" json:"vehicle_model"`
	VehicleBrand      string    `gorm:"column:vehicle_brand;comment:车辆品牌" json:"vehicle_brand"`
	VehicleColor      string    `gorm:"column:vehicle_color;comment:车身颜色" json:"vehicle_color"`
	VehicleColorCode  string    `gorm:"column:vehicle_color_code;comment:颜色代码(十六进制)" json:"vehicle_color_code"`
	SeatCount         int8      `gorm:"column:seat_count;comment:核定载人数" json:"seat_count"`
	RegisterDate      time.Time `gorm:"column:register_date;comment:车辆注册日期" json:"register_date"`
	VehicleAge        int8      `gorm:"column:vehicle_age;comment:车龄(年)" json:"vehicle_age"`
	DrivingLicenseUrl string    `gorm:"column:driving_license_url;comment:行驶证照片URL" json:"driving_license_url"`
	VehiclePhotoUrl   string    `gorm:"column:vehicle_photo_url;comment:车辆外观照片URL" json:"vehicle_photo_url"`
	Status            int8      `gorm:"column:status;comment:认证状态: 0-未认证 1-认证中 2-已认证 3-认证失败" json:"status"`
	FailReason        string    `gorm:"column:fail_reason;comment:认证失败原因" json:"fail_reason"`
	VerifyTime        time.Time `gorm:"column:verify_time;comment:认证完成时间" json:"verify_time"`
	ServiceType       int8      `gorm:"column:service_type;comment:服务类型 1-快车 2-特惠快车" json:"service_type"`
	CreatedAt         time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverVehicle) TableName() string {
	return "driver_vehicle"
}

// ----------------------------
// 6. driver_vehicle_info 司机车辆详细信息表
// ----------------------------
type DriverVehicleInfo struct {
	Id                   int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId             int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	VehicleId            int64     `gorm:"column:vehicle_id;comment:关联车辆ID" json:"vehicle_id"`
	Vin                  string    `gorm:"column:vin;comment:车辆识别代号VIN" json:"vin"`
	EngineNo             string    `gorm:"column:engine_no;comment:发动机号" json:"engine_no"`
	VehicleType          string    `gorm:"column:vehicle_type;comment:车辆类型" json:"vehicle_type"`
	Displacement         float64   `gorm:"column:displacement;comment:排量(L)" json:"displacement"`
	FuelType             string    `gorm:"column:fuel_type;comment:燃料类型: 汽油/柴油/电动/混动" json:"fuel_type"`
	InspectionExpireDate time.Time `gorm:"column:inspection_expire_date;comment:年检有效期" json:"inspection_expire_date"`
	InsuranceCompany     string    `gorm:"column:insurance_company;comment:保险公司" json:"insurance_company"`
	InsuranceExpireDate  time.Time `gorm:"column:insurance_expire_date;comment:保险有效期" json:"insurance_expire_date"`
	InsuranceUrl         string    `gorm:"column:insurance_url;comment:保险单照片URL" json:"insurance_url"`
	CreatedAt            time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverVehicleInfo) TableName() string {
	return "driver_vehicle_info"
}

// ----------------------------
// 7. driver_face 司机人脸信息表
// ----------------------------
type DriverFace struct {
	Id          int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId    int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	FaceUrl     string    `gorm:"column:face_url;comment:人脸照片URL" json:"face_url"`
	FaceFeature string    `gorm:"column:face_feature;comment:人脸特征向量(加密)" json:"face_feature"`
	Status      int8      `gorm:"column:status;comment:状态: 1-有效 2-失效" json:"status"`
	VerifyTime  time.Time `gorm:"column:verify_time;comment:人脸验证时间" json:"verify_time"`
	ExpireTime  time.Time `gorm:"column:expire_time;comment:人脸验证结果过期时间" json:"expire_time"`
	CreatedAt   time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverFace) TableName() string {
	return "driver_face"
}

// ----------------------------
// 8. driver_face_auth_log 司机人脸核验记录表
// ----------------------------
type DriverFaceAuthLog struct {
	Id         int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId   int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	AuthType   int8      `gorm:"column:auth_type;comment:核验类型: 1-出车验证 2-提现验证 3-登录验证" json:"auth_type"`
	AuthScene  string    `gorm:"column:auth_scene;comment:核验场景" json:"auth_scene"`
	FaceUrl    string    `gorm:"column:face_url;comment:人脸照片URL" json:"face_url"`
	Similarity float64   `gorm:"column:similarity;comment:相似度(%)" json:"similarity"`
	Status     int8      `gorm:"column:status;comment:核验结果: 1-通过 2-失败" json:"status"`
	FailReason string    `gorm:"column:fail_reason;comment:失败原因" json:"fail_reason"`
	DeviceId   string    `gorm:"column:device_id;comment:设备ID" json:"device_id"`
	Ip         string    `gorm:"column:ip;comment:IP地址" json:"ip"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DriverFaceAuthLog) TableName() string {
	return "driver_face_auth_log"
}

// ----------------------------
// 9. driver_status_log 司机状态变更日志表
// ----------------------------
type DriverStatusLog struct {
	Id         int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId   int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	FromStatus int8      `gorm:"column:from_status;comment:变更前状态" json:"from_status"`
	ToStatus   int8      `gorm:"column:to_status;comment:变更后状态" json:"to_status"`
	Reason     string    `gorm:"column:reason;comment:变更原因" json:"reason"`
	Operator   string    `gorm:"column:operator;comment:操作人" json:"operator"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DriverStatusLog) TableName() string {
	return "driver_status_log"
}

// ----------------------------
// 10. driver_level_config 司机等级配置表
// ----------------------------
type DriverLevelConfig struct {
	Id             int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	Level          int8      `gorm:"column:level;comment:等级: 1-5" json:"level"`
	LevelName      string    `gorm:"column:level_name;comment:等级名称" json:"level_name"`
	MinScore       float64   `gorm:"column:min_score;comment:最低服务分" json:"min_score"`
	MaxScore       float64   `gorm:"column:max_score;comment:最高服务分" json:"max_score"`
	MinOrderCount  int       `gorm:"column:min_order_count;comment:最低订单数" json:"min_order_count"`
	CommissionRate float64   `gorm:"column:commission_rate;comment:佣金比例(%)" json:"commission_rate"`
	Benefits       string    `gorm:"column:benefits;comment:等级权益(JSON)" json:"benefits"`
	Status         int8      `gorm:"column:status;comment:状态: 1-启用 2-禁用" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverLevelConfig) TableName() string {
	return "driver_level_config"
}

// ----------------------------
// 11. driver_level_record 司机等级变动记录表
// ----------------------------
type DriverLevelRecord struct {
	Id         int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId   int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	FromLevel  int8      `gorm:"column:from_level;comment:变更前等级" json:"from_level"`
	ToLevel    int8      `gorm:"column:to_level;comment:变更后等级" json:"to_level"`
	ChangeType int8      `gorm:"column:change_type;comment:变更类型: 1-升级 2-降级" json:"change_type"`
	Reason     string    `gorm:"column:reason;comment:变更原因" json:"reason"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DriverLevelRecord) TableName() string {
	return "driver_level_record"
}

// ----------------------------
// 12. driver_online_log 司机出车记录表
// ----------------------------
type DriverOnlineLog struct {
	Id             int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId       int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	OnlineTime     time.Time `gorm:"column:online_time;comment:出车时间" json:"online_time"`
	OfflineTime    time.Time `gorm:"column:offline_time;comment:收车时间" json:"offline_time"`
	OnlineDuration int       `gorm:"column:online_duration;comment:在线时长(秒)" json:"online_duration"`
	OrderCount     int       `gorm:"column:order_count;comment:完成订单数" json:"order_count"`
	Income         float64   `gorm:"column:income;comment:当日收入" json:"income"`
	StartLat       float64   `gorm:"column:start_lat;comment:起始纬度" json:"start_lat"`
	StartLng       float64   `gorm:"column:start_lng;comment:起始经度" json:"start_lng"`
	EndLat         float64   `gorm:"column:end_lat;comment:结束纬度" json:"end_lat"`
	EndLng         float64   `gorm:"column:end_lng;comment:结束经度" json:"end_lng"`
	CityId         int64     `gorm:"column:city_id;comment:服务城市ID" json:"city_id"`
	CreatedAt      time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DriverOnlineLog) TableName() string {
	return "driver_online_log"
}

// ----------------------------
// 13. driver_location_cache 司机实时位置状态表
// ----------------------------
type DriverLocationCache struct {
	Id        int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId  int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	Lat       float64   `gorm:"column:lat;comment:当前纬度" json:"lat"`
	Lng       float64   `gorm:"column:lng;comment:当前经度" json:"lng"`
	Heading   float64   `gorm:"column:heading;comment:航向角(度)" json:"heading"`
	Speed     float64   `gorm:"column:speed;comment:速度(km/h)" json:"speed"`
	Accuracy  float64   `gorm:"column:accuracy;comment:精度(米)" json:"accuracy"`
	Status    int8      `gorm:"column:status;comment:状态: 1-空车 2-有客 3-离线" json:"status"`
	OrderId   int64     `gorm:"column:order_id;comment:当前订单ID" json:"order_id"`
	CityId    int64     `gorm:"column:city_id;comment:当前城市ID" json:"city_id"`
	UpdatedAt time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverLocationCache) TableName() string {
	return "driver_location_cache"
}

// ----------------------------
// 14. dispatch_log 派单日志表
// ----------------------------
type DispatchLog struct {
	Id           int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	OrderId      int64     `gorm:"column:order_id;comment:订单ID" json:"order_id"`
	DriverId     int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	DispatchType int8      `gorm:"column:dispatch_type;comment:派单类型: 1-指派 2-抢单" json:"dispatch_type"`
	DispatchTime time.Time `gorm:"column:dispatch_time;comment:派单时间" json:"dispatch_time"`
	ExpireTime   time.Time `gorm:"column:expire_time;comment:响应截止时间" json:"expire_time"`
	Result       int8      `gorm:"column:result;comment:结果: 1-接受 2-拒绝 3-超时" json:"result"`
	ResponseTime time.Time `gorm:"column:response_time;comment:司机响应时间" json:"response_time"`
	RejectReason string    `gorm:"column:reject_reason;comment:拒绝原因" json:"reject_reason"`
	DriverLat    float64   `gorm:"column:driver_lat;comment:司机当时纬度" json:"driver_lat"`
	DriverLng    float64   `gorm:"column:driver_lng;comment:司机当时经度" json:"driver_lng"`
	Distance     int       `gorm:"column:distance;comment:距离上车点(米)" json:"distance"`
	CreatedAt    time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DispatchLog) TableName() string {
	return "dispatch_log"
}

// ----------------------------
// 15. order 订单主表
// ----------------------------
type Order struct {
	OrderId          int64      `gorm:"column:order_id;primaryKey;comment:订单ID" json:"order_id"`
	OrderNo          string     `gorm:"column:order_no;comment:订单编号" json:"order_no"`
	OrderType        int8       `gorm:"column:order_type;comment:订单类型: 1-即时单 2-预约单 3-拼车单" json:"order_type"`
	ServiceType      int8       `gorm:"column:service_type;comment:服务类型: 1-快车 2-特惠快车" json:"service_type"`
	Status           int8       `gorm:"column:status;comment:订单状态: 0-待派单 1-已派单 2-司机已接单 3-司机已到达 4-行程中 5-已完成 6-已取消 7-抢单池" json:"status"`
	DriverId         int64      `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	PassengerId      int64      `gorm:"column:passenger_id;comment:乘客ID" json:"passenger_id"`
	PassengerMobile  string     `gorm:"column:passenger_mobile;comment:乘客手机号(脱敏)" json:"passenger_mobile"`
	PassengerName    string     `gorm:"column:passenger_name;comment:乘客姓名" json:"passenger_name"`
	OriginAddress    string     `gorm:"column:origin_address;comment:起点地址" json:"origin_address"`
	OriginLat        float64    `gorm:"column:origin_lat;comment:起点纬度" json:"origin_lat"`
	OriginLng        float64    `gorm:"column:origin_lng;comment:起点经度" json:"origin_lng"`
	OriginPoi        string     `gorm:"column:origin_poi;comment:起点POI名称" json:"origin_poi"`
	DestAddress      string     `gorm:"column:dest_address;comment:终点地址" json:"dest_address"`
	DestLat          float64    `gorm:"column:dest_lat;comment:终点纬度" json:"dest_lat"`
	DestLng          float64    `gorm:"column:dest_lng;comment:终点经度" json:"dest_lng"`
	DestPoi          string     `gorm:"column:dest_poi;comment:终点POI名称" json:"dest_poi"`
	EstimateDistance int        `gorm:"column:estimate_distance;comment:预估里程(米)" json:"estimate_distance"`
	EstimateDuration int        `gorm:"column:estimate_duration;comment:预估时长(秒)" json:"estimate_duration"`
	EstimateFee      float64    `gorm:"column:estimate_fee;comment:预估费用" json:"estimate_fee"`
	ActualDistance   int        `gorm:"column:actual_distance;comment:实际里程(米)" json:"actual_distance"`
	ActualDuration   int        `gorm:"column:actual_duration;comment:实际时长(秒)" json:"actual_duration"`
	ActualFee        float64    `gorm:"column:actual_fee;comment:实际费用" json:"actual_fee"`
	BaseFee          float64    `gorm:"column:base_fee;comment:基础费用" json:"base_fee"`
	DistanceFee      float64    `gorm:"column:distance_fee;comment:里程费" json:"distance_fee"`
	DurationFee      float64    `gorm:"column:duration_fee;comment:时长费" json:"duration_fee"`
	WaitFee          float64    `gorm:"column:wait_fee;comment:等待费" json:"wait_fee"`
	DispatchAttempts int        `gorm:"column:dispatch_attempts;not null;default:0;comment:派单尝试次数" json:"dispatch_attempts"`
	RejectionCount   int        `gorm:"column:rejection_count;not null;default:0;comment:司机拒绝次数" json:"rejection_count"`
	PoolEnteredAt    *time.Time `gorm:"column:pool_entered_at;comment:进入抢单池时间" json:"pool_entered_at"`
	PoolExpireAt     *time.Time `gorm:"column:pool_expire_at;comment:抢单池超时时间" json:"pool_expire_at"`
	PoolReason       int8       `gorm:"column:pool_reason;not null;default:0;comment:进池原因 1-无司机 2-全部拒绝 3-拒绝超限 4-派单超时 5-扩半径失败" json:"pool_reason"`
	CouponId         int64      `gorm:"column:coupon_id;comment:优惠券ID" json:"coupon_id"`
	CouponAmount     float64    `gorm:"column:coupon_amount;comment:优惠券金额" json:"coupon_amount"`
	PayStatus        int8       `gorm:"column:pay_status;comment:支付状态: 1-待支付 2-已支付 3-已退款" json:"pay_status"`
	PayType          int8       `gorm:"column:pay_type;comment:支付方式: 1-微信 2-支付宝 3-余额" json:"pay_type"`
	PayTime          time.Time  `gorm:"column:pay_time;comment:支付时间" json:"pay_time"`
	CancelReason     string     `gorm:"column:cancel_reason;comment:取消原因" json:"cancel_reason"`
	CancelBy         int8       `gorm:"column:cancel_by;comment:取消方: 1-乘客 2-司机 3-系统" json:"cancel_by"`
	CancelTime       time.Time  `gorm:"column:cancel_time;comment:取消时间" json:"cancel_time"`
	CityId           int64      `gorm:"column:city_id;comment:城市ID" json:"city_id"`
	CreatedAt        time.Time  `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (Order) TableName() string {
	return "order"
}

// ----------------------------
// 16. order_evaluation 订单评价表
// ----------------------------
type OrderEvaluation struct {
	Id               int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	OrderId          int64     `gorm:"column:order_id;comment:订单ID" json:"order_id"`
	DriverId         int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	PassengerId      int64     `gorm:"column:passenger_id;comment:乘客ID" json:"passenger_id"`
	DriverScore      int8      `gorm:"column:driver_score;comment:乘客评司机分数(1-5)" json:"driver_score"`
	DriverComment    string    `gorm:"column:driver_comment;comment:乘客评司机评价内容" json:"driver_comment"`
	DriverTags       string    `gorm:"column:driver_tags;comment:乘客评司机标签(JSON数组)" json:"driver_tags"`
	PassengerScore   int8      `gorm:"column:passenger_score;comment:司机评乘客分数(1-5)" json:"passenger_score"`
	PassengerComment string    `gorm:"column:passenger_comment;comment:司机评乘客评价内容" json:"passenger_comment"`
	IsAnonymous      int8      `gorm:"column:is_anonymous;comment:是否匿名评价: 0-否 1-是" json:"is_anonymous"`
	CreatedAt        time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (OrderEvaluation) TableName() string {
	return "order_evaluation"
}

// ----------------------------
// 17. trip_service 行程服务表
// ----------------------------
type TripService struct {
	Id           int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	TripId       int64     `gorm:"column:trip_id;comment:行程ID" json:"trip_id"`
	OrderId      int64     `gorm:"column:order_id;comment:订单ID" json:"order_id"`
	DriverId     int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	PassengerId  int64     `gorm:"column:passenger_id;comment:乘客ID" json:"passenger_id"`
	AcceptTime   time.Time `gorm:"column:accept_time;comment:接单时间" json:"accept_time"`
	ArriveTime   time.Time `gorm:"column:arrive_time;comment:到达上车点时间" json:"arrive_time"`
	StartTime    time.Time `gorm:"column:start_time;comment:行程开始时间" json:"start_time"`
	EndTime      time.Time `gorm:"column:end_time;comment:行程结束时间" json:"end_time"`
	WaitDuration int       `gorm:"column:wait_duration;comment:等待乘客时长(秒)" json:"wait_duration"`
	TripDuration int       `gorm:"column:trip_duration;comment:行程时长(秒)" json:"trip_duration"`
	TripDistance int       `gorm:"column:trip_distance;comment:行程里程(米)" json:"trip_distance"`
	StartLat     float64   `gorm:"column:start_lat;comment:起点纬度" json:"start_lat"`
	StartLng     float64   `gorm:"column:start_lng;comment:起点经度" json:"start_lng"`
	EndLat       float64   `gorm:"column:end_lat;comment:终点纬度" json:"end_lat"`
	EndLng       float64   `gorm:"column:end_lng;comment:终点经度" json:"end_lng"`
	Status       int8      `gorm:"column:status;comment:状态: 1-前往上车点 2-已到达 3-行程中 4-已完成" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (TripService) TableName() string {
	return "trip_service"
}

// ----------------------------
// 18. trip_trajectory 行程轨迹归档表
// ----------------------------
type TripTrajectory struct {
	Id             int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	TripId         int64     `gorm:"column:trip_id;comment:行程ID" json:"trip_id"`
	OrderId        int64     `gorm:"column:order_id;comment:订单ID" json:"order_id"`
	DriverId       int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	TrajectoryData string    `gorm:"column:trajectory_data;comment:轨迹数据(JSON/GZIP压缩)" json:"trajectory_data"`
	PointCount     int       `gorm:"column:point_count;comment:轨迹点数量" json:"point_count"`
	StartTime      time.Time `gorm:"column:start_time;comment:轨迹开始时间" json:"start_time"`
	EndTime        time.Time `gorm:"column:end_time;comment:轨迹结束时间" json:"end_time"`
	Distance       int       `gorm:"column:distance;comment:轨迹总距离(米)" json:"distance"`
	FileUrl        string    `gorm:"column:file_url;comment:轨迹文件存储URL(大数据量时存文件)" json:"file_url"`
	CreatedAt      time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (TripTrajectory) TableName() string {
	return "trip_trajectory"
}

// ----------------------------
// 19. driver_wallet 司机钱包表
// ----------------------------
type DriverWallet struct {
	Id            int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId      int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	Balance       float64   `gorm:"column:balance;comment:可用余额" json:"balance"`
	FrozenAmount  float64   `gorm:"column:frozen_amount;comment:冻结金额" json:"frozen_amount"`
	TotalIncome   float64   `gorm:"column:total_income;comment:累计总收入" json:"total_income"`
	TotalWithdraw float64   `gorm:"column:total_withdraw;comment:累计提现金额" json:"total_withdraw"`
	Version       int       `gorm:"column:version;comment:乐观锁版本号" json:"version"`
	UpdatedAt     time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverWallet) TableName() string {
	return "driver_wallet"
}

// ----------------------------
// 20. driver_income_log 司机收入流水表
// ----------------------------
type DriverIncomeLog struct {
	Id            int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId      int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	OrderId       int64     `gorm:"column:order_id;comment:关联订单ID" json:"order_id"`
	Amount        float64   `gorm:"column:amount;comment:金额(正为收入,负为支出)" json:"amount"`
	Type          int8      `gorm:"column:type;comment:类型: 1-订单收入 2-奖励 3-罚款 4-提现 5-退款" json:"type"`
	BalanceBefore float64   `gorm:"column:balance_before;comment:变更前余额" json:"balance_before"`
	BalanceAfter  float64   `gorm:"column:balance_after;comment:变更后余额" json:"balance_after"`
	Remark        string    `gorm:"column:remark;comment:备注说明" json:"remark"`
	CreatedAt     time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DriverIncomeLog) TableName() string {
	return "driver_income_log"
}

// ----------------------------
// 21. wallet_transaction_log 钱包流水明细表
// ----------------------------
type WalletTransactionLog struct {
	Id              int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId        int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	TransactionNo   string    `gorm:"column:transaction_no;comment:流水号" json:"transaction_no"`
	TransactionType int8      `gorm:"column:transaction_type;comment:交易类型: 1-订单收入 2-奖励收入 3-提现支出 4-罚款支出 5-退款" json:"transaction_type"`
	Amount          float64   `gorm:"column:amount;comment:交易金额" json:"amount"`
	BalanceBefore   float64   `gorm:"column:balance_before;comment:交易前余额" json:"balance_before"`
	BalanceAfter    float64   `gorm:"column:balance_after;comment:交易后余额" json:"balance_after"`
	FrozenBefore    float64   `gorm:"column:frozen_before;comment:交易前冻结金额" json:"frozen_before"`
	FrozenAfter     float64   `gorm:"column:frozen_after;comment:交易后冻结金额" json:"frozen_after"`
	RelatedId       int64     `gorm:"column:related_id;comment:关联ID(订单ID/提现ID等)" json:"related_id"`
	RelatedType     string    `gorm:"column:related_type;comment:关联类型" json:"related_type"`
	Status          int8      `gorm:"column:status;comment:状态: 1-成功 2-失败 3-处理中" json:"status"`
	Remark          string    `gorm:"column:remark;comment:备注" json:"remark"`
	CreatedAt       time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (WalletTransactionLog) TableName() string {
	return "wallet_transaction_log"
}

// ----------------------------
// 22. driver_withdraw_record 司机提现记录表
// ----------------------------
type DriverWithdrawRecord struct {
	Id                int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	WithdrawNo        string    `gorm:"column:withdraw_no;comment:提现单号" json:"withdraw_no"`
	DriverId          int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	Amount            float64   `gorm:"column:amount;comment:提现金额" json:"amount"`
	Fee               float64   `gorm:"column:fee;comment:手续费" json:"fee"`
	ActualAmount      float64   `gorm:"column:actual_amount;comment:实际到账金额" json:"actual_amount"`
	BankName          string    `gorm:"column:bank_name;comment:银行名称" json:"bank_name"`
	BankCode          string    `gorm:"column:bank_code;comment:银行代码" json:"bank_code"`
	BankCardNo        string    `gorm:"column:bank_card_no;comment:银行卡号(脱敏)" json:"bank_card_no"`
	BankCardNoEncrypt string    `gorm:"column:bank_card_no_encrypt;comment:银行卡号(加密)" json:"bank_card_no_encrypt"`
	AccountName       string    `gorm:"column:account_name;comment:持卡人姓名" json:"account_name"`
	Status            int8      `gorm:"column:status;comment:状态: 1-处理中 2-成功 3-失败" json:"status"`
	FailReason        string    `gorm:"column:fail_reason;comment:失败原因" json:"fail_reason"`
	ApplyTime         time.Time `gorm:"column:apply_time;comment:申请时间" json:"apply_time"`
	FinishTime        time.Time `gorm:"column:finish_time;comment:完成时间" json:"finish_time"`
	Channel           string    `gorm:"column:channel;comment:提现渠道" json:"channel"`
	ChannelSerial     string    `gorm:"column:channel_serial;comment:渠道流水号" json:"channel_serial"`
	CreatedAt         time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverWithdrawRecord) TableName() string {
	return "driver_withdraw_record"
}

// ----------------------------
// 23. withdraw_record 提现记录表
// ----------------------------
type WithdrawRecord struct {
	Id         int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	WithdrawNo string    `gorm:"column:withdraw_no;comment:提现单号" json:"withdraw_no"`
	DriverId   int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	Amount     float64   `gorm:"column:amount;comment:提现金额" json:"amount"`
	Fee        float64   `gorm:"column:fee;comment:手续费" json:"fee"`
	BankName   string    `gorm:"column:bank_name;comment:银行名称" json:"bank_name"`
	BankCardNo string    `gorm:"column:bank_card_no;comment:银行卡号(脱敏)" json:"bank_card_no"`
	Status     int8      `gorm:"column:status;comment:状态: 1-处理中 2-成功 3-失败" json:"status"`
	FailReason string    `gorm:"column:fail_reason;comment:失败原因" json:"fail_reason"`
	ApplyTime  time.Time `gorm:"column:apply_time;comment:申请时间" json:"apply_time"`
	FinishTime time.Time `gorm:"column:finish_time;comment:完成时间" json:"finish_time"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (WithdrawRecord) TableName() string {
	return "withdraw_record"
}

// ----------------------------
// 24. service_score_log 服务分记录表
// ----------------------------
type ServiceScoreLog struct {
	Id          int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId    int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	ScoreBefore float64   `gorm:"column:score_before;comment:变更前分数" json:"score_before"`
	ScoreChange float64   `gorm:"column:score_change;comment:变更分数(正负)" json:"score_change"`
	ScoreAfter  float64   `gorm:"column:score_after;comment:变更后分数" json:"score_after"`
	ChangeType  int8      `gorm:"column:change_type;comment:变更类型: 1-好评加分 2-投诉扣分 3-违规扣分 4-系统调整" json:"change_type"`
	OrderId     int64     `gorm:"column:order_id;comment:关联订单ID" json:"order_id"`
	Remark      string    `gorm:"column:remark;comment:变更说明" json:"remark"`
	CreatedAt   time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (ServiceScoreLog) TableName() string {
	return "service_score_log"
}

// ----------------------------
// 25. driver_statistics_summary 司机统计汇总表
// ----------------------------
type DriverStatisticsSummary struct {
	Id             int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId       int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	StatDate       time.Time `gorm:"column:stat_date;comment:统计日期" json:"stat_date"`
	OnlineDuration int       `gorm:"column:online_duration;comment:在线时长(秒)" json:"online_duration"`
	OrderCount     int       `gorm:"column:order_count;comment:订单数" json:"order_count"`
	CompleteCount  int       `gorm:"column:complete_count;comment:完成订单数" json:"complete_count"`
	CancelCount    int       `gorm:"column:cancel_count;comment:取消订单数" json:"cancel_count"`
	RejectCount    int       `gorm:"column:reject_count;comment:拒绝订单数" json:"reject_count"`
	TotalIncome    float64   `gorm:"column:total_income;comment:总收入" json:"total_income"`
	OrderIncome    float64   `gorm:"column:order_income;comment:订单收入" json:"order_income"`
	BonusIncome    float64   `gorm:"column:bonus_income;comment:奖励收入" json:"bonus_income"`
	TotalDistance  int       `gorm:"column:total_distance;comment:总里程(米)" json:"total_distance"`
	TotalDuration  int       `gorm:"column:total_duration;comment:总时长(秒)" json:"total_duration"`
	AvgScore       float64   `gorm:"column:avg_score;comment:平均评分" json:"avg_score"`
	PraiseRate     float64   `gorm:"column:praise_rate;comment:好评率(%)" json:"praise_rate"`
	ComplaintCount int       `gorm:"column:complaint_count;comment:投诉数" json:"complaint_count"`
	CreatedAt      time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverStatisticsSummary) TableName() string {
	return "driver_statistics_summary"
}

// ----------------------------
// 26. pricing_rule_config 计费规则配置表
// ----------------------------
type PricingRuleConfig struct {
	Id               int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	CityId           int64     `gorm:"column:city_id;comment:城市ID" json:"city_id"`
	ServiceType      int8      `gorm:"column:service_type;comment:服务类型: 1-快车 2-特惠快车" json:"service_type"`
	RuleName         string    `gorm:"column:rule_name;comment:规则名称" json:"rule_name"`
	BasePrice        float64   `gorm:"column:base_price;comment:起步价" json:"base_price"`
	BaseDistance     int       `gorm:"column:base_distance;comment:起步里程(米)" json:"base_distance"`
	BaseDuration     int       `gorm:"column:base_duration;comment:起步时长(秒)" json:"base_duration"`
	DistancePrice    float64   `gorm:"column:distance_price;comment:里程单价(元/公里)" json:"distance_price"`
	DurationPrice    float64   `gorm:"column:duration_price;comment:时长单价(元/分钟)" json:"duration_price"`
	WaitPrice        float64   `gorm:"column:wait_price;comment:等待单价(元/分钟)" json:"wait_price"`
	WaitFreeDuration int       `gorm:"column:wait_free_duration;comment:免费等待时长(秒)" json:"wait_free_duration"`
	NightStartTime   time.Time `gorm:"column:night_start_time;comment:夜间开始时间" json:"night_start_time"`
	NightEndTime     time.Time `gorm:"column:night_end_time;comment:夜间结束时间" json:"night_end_time"`
	NightRate        float64   `gorm:"column:night_rate;comment:夜间加价倍率" json:"night_rate"`
	PeakRate         float64   `gorm:"column:peak_rate;comment:高峰加价倍率" json:"peak_rate"`
	MinPrice         float64   `gorm:"column:min_price;comment:最低消费" json:"min_price"`
	DynamicPricing   int8      `gorm:"column:dynamic_pricing;comment:是否动态调价: 0-否 1-是" json:"dynamic_pricing"`
	Status           int8      `gorm:"column:status;comment:状态: 1-启用 2-禁用" json:"status"`
	EffectiveTime    time.Time `gorm:"column:effective_time;comment:生效时间" json:"effective_time"`
	ExpireTime       time.Time `gorm:"column:expire_time;comment:失效时间" json:"expire_time"`
	CreatedAt        time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (PricingRuleConfig) TableName() string {
	return "pricing_rule_config"
}
