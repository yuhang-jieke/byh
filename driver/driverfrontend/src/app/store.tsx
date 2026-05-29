import { createContext, useContext, useState, ReactNode, useEffect, useCallback, useRef } from "react";
import * as api from "./utils/api";

export type OrderStatus =
  | "pending"
  | "dispatched"
  | "accepted"
  | "arrived"
  | "ongoing"
  | "toPay"
  | "completed"
  | "cancelled"
  | "pool";

const STATUS_MAP: Record<number, OrderStatus> = {
  0: "pending",
  1: "dispatched",
  2: "accepted",
  3: "arrived",
  4: "ongoing",
  5: "completed",
  6: "cancelled",
  7: "pool",
};

export interface Order {
  id: string;
  orderNo: string;
  passengerName: string;
  passengerPhone: string;
  from: string;
  to: string;
  fromLat: number;
  fromLng: number;
  toLat: number;
  toLng: number;
  distanceKm: number;
  estMinutes: number;
  price: number;
  carType: string;
  status: OrderStatus;
  createdAt: number;
  completedAt?: number;
  note?: string;
  driverId?: string;
  paymentMethod?: string;
  rating?: number;
  ratingComment?: string;
  ratingTags?: string[];
  nodes?: { name: string; time: number }[];
  secondsLeft?: number;
  baseFee?: number;
  distanceFee?: number;
  durationFee?: number;
  waitFee?: number;
}

export interface Driver {
  id: string;
  name: string;
  phone: string;
  plate: string;
  car: string;
  rating: number;
  online: boolean;
  listening: boolean;
  totalOrders: number;
  todayEarnings: number;
  status: "idle" | "busy" | "offline";
}

interface Store {
  orders: Order[];
  poolOrders: Order[];
  drivers: Driver[];
  currentDriverId: string;
  driverInfo: Driver | null;
  driverLoading: boolean;
  ordersLoading: boolean;
  poolOrdersLoading: boolean;
  toggleDriverStatus: () => Promise<void>;
  setDriverListening: (lat: number, lng: number) => Promise<void>;
  reportLocation: (lat: number, lng: number) => void;
  startLocationUpdates: () => void;
  stopLocationUpdates: () => void;
  acceptOrder: (orderId: number) => Promise<void>;
  rejectOrder: (orderId: number) => Promise<void>;
  cancelOrder: (orderId: number, reason: string) => Promise<void>;
  arriveOrder: (orderId: number) => Promise<void>;
  verifyPassenger: (orderId: number, phoneLast4: string) => Promise<boolean>;
  startTrip: (orderId: number) => Promise<void>;
  endTrip: (orderId: number) => Promise<void>;
  grabOrder: (orderId: number) => Promise<void>;
  loadOrders: (date?: string, cursor?: number, isAll?: boolean) => Promise<void>;
  loadDriverInfo: () => Promise<void>;
  loadPoolOrders: () => Promise<void>;
  getOrderDetail: (orderId: number) => Promise<Order | null>;
  showToast: (msg: string) => void;
  voiceMode: boolean;
  setVoiceMode: (v: boolean) => void;
}

const StoreCtx = createContext<Store | null>(null);

const DRIVER_ID = 200000001;
const SERVICE_TYPE_MAP: Record<number, string> = { 1: "快车", 2: "特惠快车" };

function parseOrderItem(item: api.ApiOrderItem): Order {
  return {
    id: String(item.orderId),
    orderNo: item.orderNo,
    passengerName: "",
    passengerPhone: "",
    from: item.originAddress,
    to: item.destAddress,
    fromLat: item.originLat ?? 0,
    fromLng: item.originLng ?? 0,
    toLat: item.destLat ?? 0,
    toLng: item.destLng ?? 0,
    distanceKm: item.distanceKm || 0,
    estMinutes: Math.round(item.durationMin || 0),
    price: item.estimateFee || 0,
    carType: SERVICE_TYPE_MAP[item.serviceType] || "快车",
    status: STATUS_MAP[item.status] || "completed",
    createdAt: item.createdAt * 1000,
    driverId: item.driverId ? String(item.driverId) : undefined,
  };
}

function parseOrderDetail(res: api.ApiGetOrderResp): Order {
  const payTypeMap: Record<number, string> = { 1: "微信", 2: "支付宝", 3: "余额" };
  return {
    id: res.orderId ? String(res.orderId) : String(res.orderNo),
    orderNo: res.orderNo,
    passengerName: res.passengerName,
    passengerPhone: res.passengerMobile,
    from: res.originAddress,
    to: res.destAddress,
    fromLat: res.originLat ?? 0,
    fromLng: res.originLng || 0,
    toLat: res.destLat || 0,
    toLng: res.destLng ?? 0,
    distanceKm: res.distanceKm || 0,
    estMinutes: Math.round((res.durationMin || 0) * 60),
    price: res.totalFee || 0,
    carType: SERVICE_TYPE_MAP[res.serviceType || 1] || "快车",
    status: STATUS_MAP[res.status] || "completed",
    createdAt: res.createdAt * 1000,
    completedAt: res.completedAt ? res.completedAt * 1000 : undefined,
    paymentMethod: res.payType ? payTypeMap[res.payType] : undefined,
    rating: res.passengerScore || undefined,
    ratingComment: res.passengerComment || undefined,
    nodes: res.nodes,
    baseFee: res.baseFee || 0,
    distanceFee: res.distanceFee || 0,
    durationFee: res.durationFee || 0,
    waitFee: res.waitFee || 0,
  };
}

function parsePoolItem(item: api.ApiPoolOrderItem): Order {
  return {
    id: String(item.orderId),
    orderNo: item.orderNo,
    passengerName: "",
    passengerPhone: "",
    from: item.originAddress,
    to: item.destAddress,
    fromLat: item.originLat ?? 0,
    fromLng: item.originLng ?? 0,
    toLat: item.destLat ?? 0,
    toLng: item.destLng ?? 0,
    distanceKm: item.estimateDistance ? item.estimateDistance / 1000 : 0,
    estMinutes: 0,
    price: item.estimateFee || 0,
    carType: SERVICE_TYPE_MAP[item.serviceType] || "快车",
    status: "pool",
    createdAt: item.createdAt * 1000,
    secondsLeft: item.secondsLeft,
  };
}

function parseDriver(d: api.ApiDriver): Driver {
  return {
    id: String(d.id),
    name: d.nickname || d.name,
    phone: d.mobile,
    plate: "-",
    car: "-",
    rating: d.serviceScore,
    online: d.workStatus > 0,
    listening: d.workStatus === 2,
    totalOrders: d.orderCount || 0,
    todayEarnings: d.totalIncome || 0,
    status: d.workStatus === 2 ? "busy" : d.workStatus === 1 ? "idle" : "offline",
  };
}

export function StoreProvider({ children }: { children: ReactNode }) {
  const [orders, setOrders] = useState<Order[]>([]);
  const [poolOrders, setPoolOrders] = useState<Order[]>([]);
  const [drivers] = useState<Driver[]>([{
    id: String(DRIVER_ID),
    name: "司机",
    phone: "",
    plate: "-",
    car: "-",
    rating: 80,
    online: false,
    listening: false,
    totalOrders: 0,
    todayEarnings: 0,
    status: "offline",
  }]);
  const [driverInfo, setDriverInfo] = useState<Driver | null>(null);
  const [driverLoading, setDriverLoading] = useState(false);
  const [ordersLoading, setOrdersLoading] = useState(false);
  const [poolOrdersLoading, setPoolOrdersLoading] = useState(false);
  const [toast, setToast] = useState<string | null>(null);
  const [voiceMode, setVoiceMode] = useState(false);
  const locationTimerRef = useRef<number | null>(null);

  const showToast = useCallback((msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 2000);
  }, []);

  // 加载司机信息
  const loadDriverInfo = useCallback(async () => {
    setDriverLoading(true);
    try {
      const res = await api.apiDriver.get(DRIVER_ID);
      const d = parseDriver(res);
      setDriverInfo(d);
      setOrders((prev) => {
        if (!prev.length) return prev;
        return prev;
      });
    } catch {
      showToast("获取司机信息失败");
    } finally {
      setDriverLoading(false);
    }
  }, [showToast]);

  // 加载订单列表
  const loadOrders = useCallback(async (date?: string, cursor?: number, isAll?: boolean) => {
    setOrdersLoading(true);
    try {
      const res = await api.apiOrderQuery.list(DRIVER_ID, { date, cursor, is_all: isAll });
      if (res.success && res.items) {
        const parsed = res.items.map(parseOrderItem);
        if (cursor) {
          // 分页加载更多时，追加到已有列表
          setOrders(prev => {
            const map = new Map<string, Order>();
            prev.forEach(o => map.set(o.id, o));
            parsed.forEach(p => map.set(p.id, p));
            return Array.from(map.values());
          });
        } else {
          // 非分页刷新时，直接用新数据替换（避免被拒订单等脏数据残留）
          setOrders(parsed);
        }
      }
    } catch {
      showToast("加载订单失败");
    } finally {
      setOrdersLoading(false);
    }
  }, [showToast]);

  // 获取订单详情
  const getOrderDetail = useCallback(async (orderId: number): Promise<Order | null> => {
    try {
      const res = await api.apiOrderQuery.detail(orderId, DRIVER_ID);
      if (res.success) return parseOrderDetail(res);
      return null;
    } catch {
      showToast("获取订单详情失败");
      return null;
    }
  }, [showToast]);

  // 加载抢单池
  const loadPoolOrders = useCallback(async () => {
    setPoolOrdersLoading(true);
    try {
      const res = await api.apiPool.list(DRIVER_ID);
      if (res.success && res.items) {
        setPoolOrders(res.items.map(parsePoolItem));
      }
    } catch {
      // 静默失败，等下次轮询
    } finally {
      setPoolOrdersLoading(false);
    }
  }, []);

  // 出车/收车/下线：根据当前状态决定操作
  const toggleDriverStatus = useCallback(async () => {
    try {
      if (driverInfo?.status === "offline") {
        // 状态 0 → 2: 出车
        await api.apiDriver.goOnline(DRIVER_ID);
        showToast("出车成功");
      } else if (driverInfo?.status === "busy") {
        // 状态 2 → 1: 收车
        await api.apiDriver.setIdle(DRIVER_ID);
        showToast("已收车");
      } else if (driverInfo?.status === "idle") {
        // 状态 1 → 0: 下线
        await api.apiDriver.goOffline(DRIVER_ID);
        showToast("已下线");
      }
      await loadDriverInfo();
    } catch (e: any) {
      showToast(e.message || "操作失败");
    }
  }, [driverInfo, loadDriverInfo, showToast]);

  // 开始听单
  const setDriverListening = useCallback(async (lat: number, lng: number) => {
    try {
      await api.apiDriver.startListening(DRIVER_ID, lat, lng);
      showToast("开始听单");
      await loadDriverInfo();
    } catch (e: any) {
      showToast(e.message || "听单失败");
    }
  }, [loadDriverInfo, showToast]);

  // 位置上报
  const reportLocation = useCallback((lat: number, lng: number) => {
    api.apiDriver.reportLocation(DRIVER_ID, lat, lng, 0, 0, 1).catch(() => {});
  }, []);

  // 启动位置上报（听单状态才上报）
  const startLocationUpdates = useCallback(() => {
    if (locationTimerRef.current) return;

    const tick = () => {
      if (!navigator.geolocation) return;
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          reportLocation(pos.coords.latitude, pos.coords.longitude);
        },
        () => {} // 失败静默处理
      );
    };

    tick(); // 立即上报一次
    locationTimerRef.current = window.setInterval(tick, 5000); // 每 5 秒上报
  }, [reportLocation]);

  // 停止位置上报
  const stopLocationUpdates = useCallback(() => {
    if (locationTimerRef.current) {
      clearInterval(locationTimerRef.current);
      locationTimerRef.current = null;
    }
  }, []);

  // 接单
  const acceptOrder = useCallback(async (orderId: number) => {
    try {
      const res = await api.apiOrder.accept(orderId, DRIVER_ID);
      if (res.success) {
        showToast("接单成功");
        await loadOrders();
      } else {
        showToast(res.message || "接单失败");
      }
    } catch (e: any) {
      showToast(e.message || "接单失败");
    }
  }, [loadOrders, showToast]);

  // 拒单
  const rejectOrder = useCallback(async (orderId: number) => {
    try {
      const res = await api.apiOrder.reject(orderId, DRIVER_ID);
      if (res.success) {
        showToast("已拒单");
        await loadOrders();
      } else {
        showToast(res.message || "拒单失败");
      }
    } catch (e: any) {
      showToast(e.message || "拒单失败");
    }
  }, [loadOrders, showToast]);

  // 取消订单
  const cancelOrder = useCallback(async (orderId: number, reason: string) => {
    try {
      const res = await api.apiOrder.cancel(orderId, DRIVER_ID, reason);
      if (res.success) {
        showToast("已取消订单");
        await loadOrders();
      } else {
        showToast(res.message || "取消失败");
      }
    } catch (e: any) {
      showToast(e.message || "取消失败");
    }
  }, [loadOrders, showToast]);

  // 到达上车点
  const arriveOrder = useCallback(async (orderId: number) => {
    try {
      const res = await api.apiOrder.arrive(orderId, DRIVER_ID);
      if (res.success) {
        showToast("已到达上车点");
        await loadOrders();
      } else {
        showToast(res.message || "操作失败");
      }
    } catch (e: any) {
      showToast(e.message || "操作失败");
    }
  }, [loadOrders, showToast]);

  // 验证乘客
  const verifyPassenger = useCallback(async (orderId: number, phoneLast4: string): Promise<boolean> => {
    try {
      const res = await api.apiOrder.verifyPassenger(orderId, DRIVER_ID, phoneLast4);
      if (res.success) {
        showToast("验证成功");
        return true;
      } else {
        showToast(res.message || "验证失败");
        return false;
      }
    } catch (e: any) {
      showToast(e.message || "验证失败");
      return false;
    }
  }, [showToast]);

  // 开始行程
  const startTrip = useCallback(async (orderId: number) => {
    try {
      const res = await api.apiOrder.startTrip(orderId, DRIVER_ID);
      if (res.success) {
        showToast("开始行程");
        await loadOrders();
      } else {
        showToast(res.message || "操作失败");
      }
    } catch (e: any) {
      showToast(e.message || "操作失败");
    }
  }, [loadOrders, showToast]);

  // 到达目的地
  const endTrip = useCallback(async (orderId: number) => {
    try {
      const res = await api.apiOrder.endTrip(orderId, DRIVER_ID);
      if (res.success) {
        showToast("行程完成");
        await loadOrders();
      } else {
        showToast(res.message || "操作失败");
      }
    } catch (e: any) {
      showToast(e.message || "操作失败");
    }
  }, [loadOrders, showToast]);

  // 抢单
  const grabOrder = useCallback(async (orderId: number) => {
    try {
      const res = await api.apiPool.grab(orderId, DRIVER_ID);
      if (res.success) {
        showToast("抢单成功");
        await loadOrders();
      } else {
        showToast(res.message || "抢单失败");
      }
    } catch (e: any) {
      showToast(e.message || "抢单失败");
    }
  }, [loadOrders, showToast]);

  // 组件卸载时停止位置上报
  useEffect(() => {
    return () => stopLocationUpdates();
  }, [stopLocationUpdates]);

  return (
    <StoreCtx.Provider value={{
      orders,
      poolOrders,
      drivers,
      currentDriverId: String(DRIVER_ID),
      driverInfo,
      driverLoading,
      ordersLoading,
      poolOrdersLoading,
      toggleDriverStatus,
      setDriverListening,
      reportLocation,
      startLocationUpdates,
      stopLocationUpdates,
      acceptOrder,
      rejectOrder,
      cancelOrder,
      arriveOrder,
      verifyPassenger,
      startTrip,
      endTrip,
      grabOrder,
      loadOrders,
      loadDriverInfo,
      loadPoolOrders,
      getOrderDetail,
      showToast,
      voiceMode,
      setVoiceMode,
    }}>
      {children}
      {/* Toast */}
      {toast && (
        <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-black/75 text-white text-xs px-4 py-2 rounded-lg z-[9999]">
          {toast}
        </div>
      )}
    </StoreCtx.Provider>
  );
}

export function useStore() {
  const s = useContext(StoreCtx);
  if (!s) throw new Error("StoreProvider missing");
  return s;
}

export const statusLabel: Record<OrderStatus, string> = {
  pending: "等待接单",
  dispatched: "待确认",
  accepted: "司机前往中",
  arrived: "司机已到达",
  ongoing: "行程中",
  toPay: "待支付",
  completed: "已完成",
  cancelled: "已取消",
  pool: "抢单池",
};

export const statusColor: Record<OrderStatus, string> = {
  pending: "bg-amber-100 text-amber-700",
  dispatched: "bg-orange-100 text-orange-700",
  accepted: "bg-blue-100 text-blue-700",
  arrived: "bg-indigo-100 text-indigo-700",
  ongoing: "bg-violet-100 text-violet-700",
  toPay: "bg-pink-100 text-pink-700",
  completed: "bg-emerald-100 text-emerald-700",
  cancelled: "bg-gray-100 text-gray-600",
  pool: "bg-orange-100 text-orange-700",
};
