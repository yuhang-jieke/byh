const BASE = "/api/v1";

// thrift 生成的 Go 结构体 JSON 标签为 PascalCase，
// 在响应层统一转为 camelCase 以匹配前端命名。
function camelizeKeys<T>(obj: T): T {
  if (obj === null || obj === undefined || typeof obj !== "object") return obj;
  if (Array.isArray(obj)) return obj.map(camelizeKeys) as unknown as T;
  const result: Record<string, any> = {};
  for (const [key, value] of Object.entries(obj as Record<string, any>)) {
    result[key.charAt(0).toLowerCase() + key.slice(1)] = camelizeKeys(value);
  }
  return result as T;
}

async function get<T = any>(path: string, query?: Record<string, string>): Promise<T> {
  const url = query ? `${path}?${new URLSearchParams(query)}` : path;
  const res = await fetch(`${BASE}${url}`);
  if (!res.ok) {
    const body = await res.text().catch(() => "");
    throw new Error(`HTTP ${res.status}: ${body}`);
  }
  return camelizeKeys(await res.json()) as T;
}

async function post<T = any>(path: string, body?: Record<string, any>): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const body = await res.text().catch(() => "");
    throw new Error(`HTTP ${res.status}: ${body}`);
  }
  return camelizeKeys(await res.json()) as T;
}

// ---------- 基础接口 ----------

export interface ApiOrder {
  success: boolean;
  orderId?: number;
  orderNo: string;
  status: number;
  createdAt: number;
  completedAt?: number;
  originAddress: string;
  destAddress: string;
  distanceKm?: number;
  durationMin?: number;
  passengerName: string;
  passengerMobile: string;
  passengerScore?: number;
  passengerComment?: string;
  totalFee?: number;
  platformCommission?: number;
  driverIncome?: number;
  baseFee?: number;
  distanceFee?: number;
  durationFee?: number;
  waitFee?: number;
  payType?: number;
  serviceType?: number;
  nodes?: { name: string; time: number }[];
  originLat?: number;
  originLng?: number;
  destLat?: number;
  destLng?: number;
}

export interface ApiDriver {
  id: number;
  name: string;
  mobile: string;
  nickname: string;
  workStatus: number;
  serviceScore: number;
  orderCount: number;
  totalIncome: number;
}

export interface ApiOrderItem {
  orderId: number;
  orderNo: string;
  serviceType: number;
  originAddress: string;
  destAddress: string;
  distanceKm: number;
  durationMin: number;
  status: number;
  createdAt: number;
  estimateFee: number;
  driverId: number;
  originLat: number;
  originLng: number;
  destLat: number;
  destLng: number;
}

export interface ApiListOrdersResp {
  success: boolean;
  items: ApiOrderItem[];
}

export interface ApiPoolOrderItem {
  orderId: number;
  orderNo: string;
  serviceType: number;
  originLat: number;
  originLng: number;
  originAddress: string;
  destLat: number;
  destLng: number;
  destAddress: string;
  estimateDistance: number;
  estimateFee: number;
  createdAt: number;
  secondsLeft: number;
}

export interface ApiPoolListResp {
  success: boolean;
  items: ApiPoolOrderItem[];
  total: number;
}

export interface ApiGetOrderResp extends ApiOrder {}

export interface ApiRoutePlanResp {
  distance: number;   // 米
  duration: number;   // 秒
  polyline: { lng: number; lat: number }[];
}

// ---------- 司机相关 ----------

export const apiDriver = {
  get(id: number) {
    return get<ApiDriver>(`/drivers/${id}`);
  },
  goOnline(id: number) {
    return post(`/drivers/${id}/go-online`);
  },
  setIdle(id: number) {
    return post(`/drivers/${id}/set-idle`);
  },
  goOffline(id: number) {
    return post(`/drivers/${id}/go-offline`);
  },
  startListening(id: number, lat: number, lng: number) {
    return post(`/drivers/${id}/start-listening`, { lat, lng });
  },
  reportLocation(id: number, lat: number, lng: number, heading: number, speed: number, status: number) {
    return post(`/drivers/${id}/report-location`, { lat, lng, heading, speed, status });
  },
};

// ---------- 订单相关 ----------

export const apiOrder = {
  dispatch(body: { order_id: number; service_type: number; origin_lat: number; origin_lng: number; passenger_id: number }) {
    return post("/orders/dispatch", body);
  },
  accept(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/accept`, { driver_id: driverId });
  },
  reject(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/reject`, { driver_id: driverId });
  },
  cancel(orderId: number, driverId: number, cancel_reason: string) {
    return post(`/orders/${orderId}/cancel`, { driver_id: driverId, cancel_reason });
  },
  arrive(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/arrive`, { driver_id: driverId });
  },
  verifyPassenger(orderId: number, driverId: number, phone_last4: string) {
    return post(`/orders/${orderId}/verify-passenger`, { driver_id: driverId, phone_last4 });
  },
  startTrip(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/start-trip`, { driver_id: driverId });
  },
  endTrip(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/end-trip`, { driver_id: driverId });
  },
};

// ---------- 订单查询 ----------

export const apiOrderQuery = {
  list(driverId: number, params?: { date?: string; cursor?: number; is_all?: boolean }) {
    const q: Record<string, string> = {};
    if (params?.date) q.date = params.date;
    if (params?.cursor) q.cursor = String(params.cursor);
    if (params?.is_all) q.is_all = "true";
    return get<ApiListOrdersResp>(`/drivers/${driverId}/orders`, q);
  },
  detail(orderId: number, driverId: number) {
    return get<ApiGetOrderResp>(`/orders/${orderId}/detail`, { driver_id: String(driverId) });
  },
};

// ---------- 抢单池 ----------

export const apiPool = {
  list(driverId: number, page = 1, pageSize = 20) {
    return post("/pool/list", { driver_id: driverId, page, page_size: pageSize });
  },
  grab(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/grab`, { driver_id: driverId });
  },
};

// ---------- 路径规划 ----------

export const apiRoute = {
  plan(origin_lat: number, origin_lng: number, dest_lat: number, dest_lng: number) {
    return post<ApiRoutePlanResp>("/route/plan", { origin_lat, origin_lng, dest_lat, dest_lng });
  },
};
