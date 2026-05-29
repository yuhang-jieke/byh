namespace go driver

struct CreateDriverReq {
    1: string Name
}

struct CreateDriverResp {
    1: i64 Id
}

struct GetDriverReq {
    1: i64 Id
}

struct GetDriverResp {
    1: i64 Id
    2: string Name
    3: i32 Status
    4: string Mobile
    5: string Nickname
    6: i32 WorkStatus
    7: double ServiceScore
    8: i32 OrderCount
    9: double TotalIncome
}

struct ListDriverReq {
}

struct DriverItem {
    1: i64 Id
    2: string Name
    3: i32 Status
}

struct ListDriverResp {
    1: list<DriverItem> Items
}

struct UpdateDriverReq {
    1: i64 Id
    2: string Name
}

struct UpdateDriverResp {
    1: bool Success
}

struct DeleteDriverReq {
    1: i64 Id
}

struct DeleteDriverResp {
    1: bool Success
}

struct GoOnlineReq {
    1: i64 DriverId
}

struct GoOnlineResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct SetIdleReq {
    1: i64 DriverId
}

struct SetIdleResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct StartListeningReq {
    1: i64 DriverId
    2: double Lat
    3: double Lng
}

struct StartListeningResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct GoOfflineReq {
    1: i64 DriverId
}

struct GoOfflineResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct ReportLocationReq {
    1: i64 DriverId
    2: double Lat
    3: double Lng
    4: double Heading
    5: double Speed
    6: i32 Status
}

struct ReportLocationResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct DispatchOrderReq {
    1: i64 OrderId
    2: string OrderNo
    3: i32 ServiceType
    4: double OriginLat
    5: double OriginLng
    6: i64 PassengerId
}

struct DispatchOrderResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct AcceptOrderReq {
    1: i64 OrderId
    2: i64 DriverId
}

struct AcceptOrderResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct RejectOrderReq {
    1: i64 OrderId
    2: i64 DriverId
}

struct RejectOrderResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct CancelOrderReq {
    1: i64 OrderId
    2: i64 DriverId
    3: string CancelReason
}

struct CancelOrderResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct DriverArriveReq {
    1: i64 OrderId
    2: i64 DriverId
}

struct DriverArriveResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct StartTripReq {
    1: i64 OrderId
    2: i64 DriverId
}

struct StartTripResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct EndTripReq {
    1: i64 OrderId
    2: i64 DriverId
}

struct EndTripResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct VerifyPassengerPhoneReq {
    1: i64 OrderId
    2: i64 DriverId
    3: string PhoneLast4
}

struct VerifyPassengerPhoneResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
}

struct ListPoolOrdersReq {
    1: i64 DriverId
    2: i32 Page
    3: i32 PageSize
}

struct PoolOrderItem {
    1: i64 OrderId
    2: string OrderNo
    3: i32 ServiceType
    4: double OriginLat
    5: double OriginLng
    6: string OriginAddress
    7: double DestLat
    8: double DestLng
    9: string DestAddress
    10: i64 PassengerId
    11: string PassengerName
    12: double EstimateDistance
    13: double EstimateFee
    14: i64 CreatedAt
    15: i64 SecondsLeft
}

struct ListPoolOrdersResp {
    1: bool Success
    2: list<PoolOrderItem> Items
    3: i32 Total
}

struct GrabOrderReq {
    1: i64 OrderId
    2: i64 DriverId
}

struct GrabOrderResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
    4: i64 OrderId
}

struct ListOrdersReq {
    1: i64 DriverId
    2: string Date
    3: i32 Cursor
    4: bool IsAll
}

struct OrderItem {
    1: i64 OrderId
    2: string OrderNo
    3: i32 ServiceType
    4: string OriginAddress
    5: string DestAddress
    6: double DistanceKm
    7: double DurationMin
    8: i32 Status
    9: i64 CreatedAt
    10: double EstimateFee
    11: i64 DriverId
    12: double OriginLat
    13: double OriginLng
    14: double DestLat
    15: double DestLng
}

struct ListOrdersResp {
    1: bool Success
    2: list<OrderItem> Items
}

struct GetOrderReq {
    1: i64 OrderId
    2: i64 DriverId
}

struct OrderNode {
    1: string Name
    2: i64 Time
}

struct GetOrderResp {
    1: bool Success
    2: i32 ErrCode
    3: string Message
    25: i64 OrderId
    4: string OrderNo
    5: i32 Status
    6: i64 CreatedAt
    7: i64 CompletedAt
    8: string OriginAddress
    9: string DestAddress
    10: double DistanceKm
    11: double DurationMin
    12: string PassengerName
    13: string PassengerMobile
    14: i32 PassengerScore
    15: string PassengerComment
    16: double TotalFee
    17: double PlatformCommission
    18: double DriverIncome
    19: i32 PayType
    20: list<OrderNode> Nodes
    21: double BaseFee
    22: double DistanceFee
    23: double DurationFee
    24: double WaitFee
    26: double OriginLat
    27: double OriginLng
    28: double DestLat
    29: double DestLng
}

struct GetTodayStatsReq {
    1: i64 DriverId
    2: string Date
}

struct TodayStatsResp {
    1: bool Success
    2: i32 CompletedOrders
    3: double TotalEarnings
    4: i32 OnlineSeconds
    5: double TotalKm
}

struct GetCurrentOrderReq {
    1: i64 DriverId
}

struct CurrentOrderResp {
    1: bool Success
    2: i64 OrderId
    3: string OrderNo
    4: i32 Status
    5: string OriginAddress
    6: double OriginLat
    7: double OriginLng
    8: string DestAddress
    9: double DestLat
    10: double DestLng
    11: string PassengerName
    12: string PassengerMobile
    13: double EstimateFee
}

struct GetWalletReq {
    1: i64 DriverId
}

struct WalletResp {
    1: bool Success
    2: double Balance
    3: double TotalIncome
    4: double FrozenAmount
}

service DriverService {
    CreateDriverResp Create(1: CreateDriverReq req)
    GetDriverResp Get(1: GetDriverReq req)
    ListDriverResp List(1: ListDriverReq req)
    UpdateDriverResp Update(1: UpdateDriverReq req)
    DeleteDriverResp Delete(1: DeleteDriverReq req)
    GoOnlineResp GoOnline(1: GoOnlineReq req)
    SetIdleResp SetIdle(1: SetIdleReq req)
    StartListeningResp StartListening(1: StartListeningReq req)
    GoOfflineResp GoOffline(1: GoOfflineReq req)
    ReportLocationResp ReportLocation(1: ReportLocationReq req)
    DispatchOrderResp DispatchOrder(1: DispatchOrderReq req)
    AcceptOrderResp AcceptOrder(1: AcceptOrderReq req)
    RejectOrderResp RejectOrder(1: RejectOrderReq req)
    CancelOrderResp CancelOrder(1: CancelOrderReq req)
    DriverArriveResp DriverArrive(1: DriverArriveReq req)
    StartTripResp StartTrip(1: StartTripReq req)
    EndTripResp EndTrip(1: EndTripReq req)
    VerifyPassengerPhoneResp VerifyPassengerPhone(1: VerifyPassengerPhoneReq req)
    ListPoolOrdersResp ListPoolOrders(1: ListPoolOrdersReq req)
    GrabOrderResp GrabOrder(1: GrabOrderReq req)
    ListOrdersResp ListOrders(1: ListOrdersReq req)
    GetOrderResp GetOrder(1: GetOrderReq req)
    TodayStatsResp GetTodayStats(1: GetTodayStatsReq req)
    CurrentOrderResp GetCurrentOrder(1: GetCurrentOrderReq req)
    WalletResp GetWallet(1: GetWalletReq req)
}
