import AMapLoader from "@amap/amap-jsapi-loader";

// 高德地图配置 - 使用"花小猪"Key (Web 端)
const AMap_KEY = "3207c85d5ebe03ec3e14a94971d44fea";
const AMap_SECURITY_CODE = "501d9f1049ccf38f8a0aaefac576b1c4";

// 确保在加载地图前设置安全码
if (typeof window !== "undefined") {
  (window as any)._AMapSecurity = { securityJsCode: AMap_SECURITY_CODE };
}

let _AMap: any = null;
let _loadPromise: Promise<any> | null = null;

export async function loadAMap() {
  if (_AMap) return _AMap;
  if (!_loadPromise) {
    _loadPromise = AMapLoader.load({
      key: AMap_KEY,
      version: "2.0",
      plugins: ["AMap.Geocoder", "AMap.AutoComplete", "AMap.Geolocation", "AMap.Driving"],
    });
    _loadPromise.then((AMap) => { _AMap = AMap; });
  }
  return _loadPromise;
}

export function createMap(container: HTMLElement, opts?: any) {
  const AMap = _AMap;
  return new AMap.Map(container, {
    zoom: 13,
    center: [118.3, 33.95],
    viewMode: "2D",
    ...opts,
  });
}

// 坐标格式统一为 [lat, lng] 传入，内部转换为 [lng, lat]
export async function geocode(address: string, city = "全国"): Promise<[number, number] | null> {
  console.log(`[高德地理编码] 开始: "${address}"`);
  const AMap = await loadAMap();
  return new Promise((resolve) => {
    let settled = false;
    const timer = setTimeout(() => {
      if (!settled) {
        settled = true;
        console.warn(`[高德地理编码] 超时: "${address}"`);
        resolve(null);
      }
    }, 5000);
    const geocoder = new AMap.Geocoder({ city, radius: 1000, extensions: "base" });
    geocoder.getLocation(address, (status: string, result: any) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      console.log(`[高德地理编码] 回调: status=${status}, count=${result?.geocodes?.length || 0}`);
      if (status === "complete" && result.geocodes?.length > 0) {
        const loc = result.geocodes[0].location;
        resolve([loc.getLat(), loc.getLng()]);
      } else {
        console.warn(`[高德地理编码] 失败: "${address}", status=${status}`);
        resolve(null);
      }
    });
  });
}

export async function reverseGeocode(lat: number, lng: number): Promise<string> {
  const AMap = await loadAMap();
  return new Promise((resolve) => {
    const geocoder = new AMap.Geocoder({ city: "全国", radius: 1000, extensions: "base" });
    geocoder.getAddress([lng, lat], (status: string, result: any) => {
      if (status === "complete" && result.regeocode) {
        resolve(result.regeocode.formattedAddress || result.regeocode.sematicDescription || "");
      } else {
        resolve("");
      }
    });
  });
}

// 规划驾车路线（优先 BFF API，降级 JS SDK）
// 返回格式兼容 AmapView.tsx 的调用：route.paths[0].steps.map(s => s.polyline).flat()
export async function searchDrivingRoute(
  origin: [number, number],
  destination: [number, number]
): Promise<{
  distance: number;
  time: number;
  paths: { steps: { polyline: [number, number][] }[] }[];
}> {
  // 优先调用 BFF API（已修复数字签名）
  try {
    const { apiRoute } = await import('./api');
    const resp = await apiRoute.plan(origin[0], origin[1], destination[0], destination[1]);
    if (resp && resp.polyline && resp.polyline.length > 0) {
      console.log(`[路线规划] BFF 规划成功，距离：${resp.distance}米，预计：${resp.duration}秒`);
      return {
        distance: resp.distance,
        time: resp.duration,
        paths: [{ steps: [{ polyline: resp.polyline.map(p => [p.lng, p.lat]) }] }]
      };
    }
    console.warn('[路线规划] BFF 返回空');
  } catch (e) {
    console.warn('[路线规划] BFF 失败，降级 JS SDK:', e);
  }

  // 降级：使用 JS SDK
  try {
    const AMap = await loadAMap();
    const driving = new AMap.Driving({ policy: AMap.DrivingPolicy.LEAST_TIME });

    return new Promise((resolve) => {
      driving.search(
        new AMap.LngLat(origin[1], origin[0]),
        new AMap.LngLat(destination[1], destination[0]),
        (status: string, result: any) => {
          if (status !== "complete" || !result.routes || result.routes.length === 0) {
            console.warn(`[路线规划] JS SDK 无路线`, status, result?.info);
            resolve({ distance: 0, time: 0, paths: [] });
            return;
          }

          const route = result.routes[0];
          const steps = route.steps.map((step: any) => ({
            polyline: step.path.map((p: any) => [p.lng, p.lat])
          }));

          resolve({
            distance: route.distance,
            time: route.time,
            paths: [{ steps }]
          });
        }
      );
    });
  } catch (e) {
    console.error('[路线规划] JS SDK 调用失败:', e);
    return { distance: 0, time: 0, paths: [] };
  }
}

// 直接获取路径规划坐标数组（兼容 effect #6 的简化调用）
export async function planDrivingRoute(
  fromLat: number, fromLng: number, toLat: number, toLng: number
): Promise<{ distance: number; duration: number; polyline: { lng: number; lat: number }[] }> {
  const route = await searchDrivingRoute([fromLat, fromLng], [toLat, toLng]);
  if (!route.paths?.length) {
    return { distance: 0, duration: 0, polyline: [] };
  }
  const points = route.paths[0].steps.map(s => s.polyline).flat();
  return {
    distance: route.distance,
    duration: route.time,
    polyline: points.map(([lng, lat]: number[]) => ({ lng, lat })),
  };
}
