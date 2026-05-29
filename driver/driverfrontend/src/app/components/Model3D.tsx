import { Suspense, useRef, useEffect, useState, useCallback } from "react";
import { Canvas, useFrame } from "@react-three/fiber";
import { OrbitControls, useGLTF } from "@react-three/drei";
import * as THREE from "three";

/* ============================================================
   模型场景
   ============================================================ */
function ModelScene({ state }: { state: string }) {
  const groupRef = useRef<THREE.Group>(null);
  const groundInnerRef = useRef<THREE.Mesh>(null);
  const groundOuterRef = useRef<THREE.Mesh>(null);
  const { scene } = useGLTF("/dh/model.glb");

  const modelRef = useRef<THREE.Group | null>(null);
  if (!modelRef.current && scene) {
    modelRef.current = scene.clone(true);
  }

  // 居中
  useEffect(() => {
    if (!modelRef.current) return;
    const box = new THREE.Box3().setFromObject(modelRef.current);
    const center = box.getCenter(new THREE.Vector3());
    modelRef.current.position.set(-center.x, -center.y, -center.z);
  }, []);

  // ---------- 交互状态（纯 ref，零重渲染） ----------
  const mouseVel = useRef({ x: 0, y: 0 });
  const mouseRot = useRef({ x: 0, y: 0 });
  const clickTimeRef = useRef(0);
  const entryTimeRef = useRef(performance.now());

  const lastFrameTime = useRef(0);

  useFrame(({ clock, pointer }) => {
    if (!groupRef.current) return;

    const now = clock.getElapsedTime();
    if (now - lastFrameTime.current < 1 / 30) return;
    lastFrameTime.current = now;

    // ===== 1) 入场弹性 =====
    const entryElapsed = (performance.now() - entryTimeRef.current) / 1000;
    const entryP = Math.min(entryElapsed / 0.8, 1);
    // easeOutBack：先超调 12% 再回弹
    const c1 = 1.7;
    const c3 = c1 + 1;
    const overshoot = 1 + c3 * Math.pow(entryP - 1, 3) + c1 * Math.pow(entryP - 1, 2);
    const entryOffset = -0.35 * (1 - overshoot);
    const entryScale = 0.92 + 0.08 * overshoot;

    // ===== 2) 鼠标跟随（带惯性皮筋效果） =====
    const targetRx = -pointer.y * 0.06;
    const targetRy = pointer.x * 0.08;
    mouseVel.current.x += (targetRx - mouseRot.current.x) * 0.005;
    mouseVel.current.y += (targetRy - mouseRot.current.y) * 0.005;
    mouseVel.current.x *= 0.88;
    mouseVel.current.y *= 0.88;
    mouseRot.current.x += mouseVel.current.x;
    mouseRot.current.y += mouseVel.current.y;

    // ===== 3) 点击挤压拉伸 =====
    const clickMs = performance.now() - clickTimeRef.current;
    let sx = 1, sy = 1;
    if (clickMs < 500) {
      const p = clickMs / 500;
      const decay = 1 - p;
      const wave = Math.sin(p * Math.PI * 2) * 0.06 * decay;
      sy = 1 + wave;           // Y 轴：>1 拉伸，<1 挤压
      sx = 1 - wave * 0.4;     // XZ 轴：反向补偿，体积感
    }

    // ===== 4) 双层地面呼吸光晕 =====
    if (groundInnerRef.current) {
      const mat = groundInnerRef.current.material as THREE.MeshBasicMaterial;
      const breath = 0.5 + 0.5 * Math.sin(now * 1.8);
      mat.opacity = 0.2 + 0.25 * breath;
      const s = 1 + 0.08 * breath;
      groundInnerRef.current.scale.set(s, s, 1);
    }
    if (groundOuterRef.current) {
      const mat = groundOuterRef.current.material as THREE.MeshBasicMaterial;
      const slowBreath = 0.5 + 0.5 * Math.sin(now * 0.7 + 1.2);
      mat.opacity = 0.06 + 0.08 * slowBreath;
      const s = 1 + 0.12 * slowBreath;
      groundOuterRef.current.scale.set(s, s, 1);
    }

    // ===== 5) 状态基础动画 =====
    const floatY = Math.sin(now * 1.2) * 0.025;

    const stateRy =
      state === "listening" ? Math.sin(now * 0.6) * 0.15
      : state === "thinking" ? now * 0.15
      : state === "speaking" ? Math.sin(now * 1.5) * 0.08
      : Math.sin(now * 0.3) * 0.12;

    const stateRx =
      state === "listening" ? 0.03
      : state === "thinking" ? -0.02
      : state === "speaking" ? Math.sin(now * 4) * 0.02
      : Math.sin(now * 0.5) * 0.02;

    const stateY =
      state === "listening" ? floatY
      : state === "thinking" ? floatY - 0.02
      : state === "speaking" ? floatY + Math.sin(now * 6) * 0.008
      : floatY;

    // ===== 合成最终变换 =====
    groupRef.current.position.y = stateY + entryOffset;
    groupRef.current.rotation.x = stateRx + mouseRot.current.x;
    groupRef.current.rotation.y = stateRy + mouseRot.current.y;
    groupRef.current.scale.set(0.7 * entryScale * sx, 0.7 * entryScale * sy, 0.7 * entryScale * sx);
  });

  const handlePointerDown = useCallback(() => {
    clickTimeRef.current = performance.now();
  }, []);

  // 状态发光
  useEffect(() => {
    if (!modelRef.current) return;
    const cfg: Record<string, [string, number]> = {
      listening: ["#B0E0E6", 0.5],
      speaking: ["#f472b6", 0.25],
      thinking: ["#B0E0E6", 0.3],
      idle: ["#000000", 0],
    };
    const [hex, intensity] = cfg[state] || cfg.idle;
    const emColor = new THREE.Color(hex);

    modelRef.current.traverse((child) => {
      if (!child.isMesh) return;
      const mesh = child as THREE.Mesh;
      const mat = mesh.material as THREE.MeshStandardMaterial;
      if (!mat) return;
      mat.emissive = emColor;
      mat.emissiveIntensity = intensity;
    });
  }, [state]);

  if (!modelRef.current) return null;

  return (
    <group ref={groupRef} onPointerDown={handlePointerDown}>
      <primitive object={modelRef.current} />

      {/* 内层光晕 — 快速呼吸 */}
      <mesh ref={groundInnerRef} rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.78, 0]}>
        <ringGeometry args={[0.25, 0.5, 32]} />
        <meshBasicMaterial color="#fbcfe8" transparent opacity={0.25} depthWrite={false} />
      </mesh>

      {/* 外层光晕 — 慢速扩散 */}
      <mesh ref={groundOuterRef} rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.78, 0]}>
        <ringGeometry args={[0.5, 0.85, 32]} />
        <meshBasicMaterial color="#bae6fd" transparent opacity={0.1} depthWrite={false} />
      </mesh>
    </group>
  );
}

/* ============================================================
   加载占位
   ============================================================ */
function ModelLoader() {
  return (
    <div className="w-full h-full flex items-center justify-center">
      <div className="flex flex-col items-center gap-2">
        <div className="w-8 h-8 border-2 border-pink-300 border-t-pink-500 rounded-full animate-spin" />
        <span className="text-xs text-gray-400">数字人加载中...</span>
      </div>
    </div>
  );
}

/* ============================================================
   3D 模型容器
   ============================================================ */
export default function ModelViewer({
  state,
  onReady,
  onError,
}: {
  state: string;
  onReady?: () => void;
  onError?: () => void;
}) {
  const [phase, setPhase] = useState<"checking" | "ready" | "failed">("checking");

  useEffect(() => {
    let cancelled = false;
    const timer = setTimeout(() => {
      if (!cancelled) {
        setPhase("failed");
        onError?.();
      }
    }, 8000);

    fetch("/dh/model.glb", { method: "HEAD" })
      .then((res) => {
        if (cancelled) return;
        clearTimeout(timer);
        if (res.ok) {
          setPhase("ready");
          onReady?.();
        } else {
          setPhase("failed");
          onError?.();
        }
      })
      .catch(() => {
        if (cancelled) return;
        clearTimeout(timer);
        setPhase("failed");
        onError?.();
      });

    return () => { cancelled = true; clearTimeout(timer); };
  }, []);

  if (phase === "failed") return null;

  return (
    <Suspense fallback={<ModelLoader />}>
      <Canvas
        camera={{ position: [0, 0.6, 3.8], fov: 32, near: 0.1, far: 10 }}
        gl={{ antialias: true, alpha: true }}
        dpr={1}
        shadows={false}
        style={{ width: "100%", height: "100%" }}
      >
        <ambientLight intensity={0.5} />
        <directionalLight position={[3, 4, 5]} intensity={0.8} />
        <directionalLight position={[-2, 1, 3]} intensity={0.4} color="#93c5fd" />
        <pointLight position={[0, -1, 3]} intensity={0.2} color="#fbcfe8" />
        <hemisphereLight args={["#fce4ec", "#e8e8e8", 0.4]} />

        <ModelScene state={state} />

        <OrbitControls
          enableZoom={false}
          enablePan={false}
          minPolarAngle={Math.PI / 3}
          maxPolarAngle={Math.PI / 1.6}
          rotateSpeed={0.6}
        />

        <fog attach="fog" args={["#fce4ec", 3, 6]} />
      </Canvas>
    </Suspense>
  );
}
