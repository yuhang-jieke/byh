import { useState, useCallback, useRef, useEffect } from "react";
import { createSpeechEngine, isSpeechSupported, SpeechEngine } from "./speechEngine";
import { useStore } from "../store";

export type DhState = "idle" | "listening" | "thinking" | "speaking";

export interface Message {
  role: "user" | "assistant";
  content: string;
}

const DH_STORAGE_KEY = "dh_messages";
const DH_FIRST_TIME_KEY = "dh_first_time";

// ---------- 回复匹配引擎 ----------

function formatStatus(status: string): string {
  const map: Record<string, string> = {
    offline: "当前已下线",
    idle: "当前空闲在线",
    busy: "当前忙碌中",
  };
  return map[status] || status;
}

function generateReply(input: string, ctx: {
  driverStatus: string;
  isListening: boolean;
  totalOrders: number;
  todayEarnings: number;
}): string {
  const text = input.trim().toLowerCase();

  if (/你好|嗨|在吗|hello|hi/i.test(text)) {
    return "你好呀！我是花小猪助手，有什么可以帮你的吗？";
  }

  if (/几单|多少单|接单情况|今天怎么样|接了多少|收入|赚了/i.test(text)) {
    const statusText = ctx.driverStatus === "offline" ? "今天还没出车" : `目前${formatStatus(ctx.driverStatus)}`;
    return `你今天接了 ${ctx.totalOrders} 单，总收入 ${ctx.todayEarnings} 元，${statusText}。`;
  }

  if (/什么状态|在干嘛|在做什么|忙不忙|现在.*状态/i.test(text)) {
    const listeningText = ctx.isListening ? "正在听单中" : "没有在听单";
    return `你目前 ${formatStatus(ctx.driverStatus)}，${listeningText}。`;
  }

  if (/你叫什么|你是谁|名字/i.test(text)) {
    return "我是花小猪助手，你的专属 AI 驾驶伙伴！";
  }

  if (/可爱|好看|漂亮|乖/i.test(text)) {
    return "嘻嘻，谢谢！穿上这身机甲，我可是要保护你安全驾驶的！";
  }

  if (/再见|拜拜|退下|bye|先不聊/i.test(text)) {
    return "好的，有事随时叫我，我在右下角等你～";
  }

  if (/谢谢|thank/i.test(text)) {
    return "不客气！有需要随时找我～";
  }

  if (/无聊|聊天|讲个笑话|笑话/i.test(text)) {
    return "那我给你讲个笑话：为什么程序员总是分不清万圣节和圣诞节？...因为 Oct 31 = Dec 25！";
  }

  return "这个问题我还不太明白，不过我会继续学习的！你可以问我接单情况或者跟我聊聊天～";
}

// 调后端 AI API，失败时回退到本地匹配
async function fetchAIReply(text: string, driverId: number, ctx: {
  driverStatus: string;
  isListening: boolean;
  totalOrders: number;
  todayEarnings: number;
  activeOrderStatus: string | null;
}): Promise<string> {
  try {
    const res = await fetch("/api/v1/ai/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        message: text,
        driver_id: driverId,
      }),
    });
    if (!res.ok) throw new Error("API error");
    const data = await res.json();
    if (data.reply) return data.reply;
  } catch {}
  return generateReply(text, ctx);
}

// ---------- 对话持久化 ----------

function loadHistory(): Message[] {
  try {
    const raw = localStorage.getItem(DH_STORAGE_KEY);
    if (raw) return JSON.parse(raw);
  } catch {}
  return [];
}

function saveHistory(messages: Message[]) {
  try {
    // 只保留最近 20 条
    localStorage.setItem(DH_STORAGE_KEY, JSON.stringify(messages.slice(-20)));
  } catch {}
}

function isFirstTime(): boolean {
  return !localStorage.getItem(DH_FIRST_TIME_KEY);
}

function markFirstTimeDone() {
  localStorage.setItem(DH_FIRST_TIME_KEY, "true");
}

// ---------- Hook ----------

export function useDigitalHuman() {
  const store = useStore();
  const [state, setState] = useState<DhState>("idle");
  const [messages, setMessages] = useState<Message[]>(() => {
    const history = loadHistory();
    // 如果有历史，不追加欢迎语。通过 showWelcome 控制
    return history;
  });
  const [showWelcome, setShowWelcome] = useState(() => messages.length === 0);
  const [volumeLevel, setVolumeLevel] = useState(0);
  const [micGranted, setMicGranted] = useState<boolean | null>(null);
  const [speechSupported] = useState(isSpeechSupported);
  const [voiceMode, setVoiceMode] = useState(false);
  const [voicePer, setVoicePer] = useState("3");
  const engineRef = useRef<SpeechEngine | null>(null);
  const lastInputRef = useRef("");
  const lastInputTimeRef = useRef(0);

  // 构造上下文
  const getContext = useCallback(() => {
    const di = store.driverInfo;
    const activeOrder = store.orders.find(o =>
      ["accepted", "arrived", "ongoing"].includes(o.status)
    );
    return {
      driverStatus: di?.status || "offline",
      isListening: di?.listening || false,
      totalOrders: di?.totalOrders || 0,
      todayEarnings: di?.todayEarnings || 0,
      activeOrderStatus: activeOrder?.status || null,
    };
  }, [store.driverInfo, store.orders]);

  // 添加消息
  const addMessage = useCallback((msg: Message) => {
    setMessages(prev => {
      const next = [...prev, msg];
      saveHistory(next);
      return next;
    });
  }, []);

  // 处理用户输入
  const handleUserInput = useCallback(async (text: string) => {
    const trimmed = text.trim();
    if (!trimmed) return;

    const now = Date.now();
    if (trimmed === lastInputRef.current && now - lastInputTimeRef.current < 5000) {
      return;
    }
    lastInputRef.current = trimmed;
    lastInputTimeRef.current = now;

    setShowWelcome(false);
    addMessage({ role: "user", content: trimmed });
    setState("thinking");

    const reply = await fetchAIReply(trimmed, 200000001, getContext());
    addMessage({ role: "assistant", content: reply });

    if (voiceMode && engineRef.current) {
      engineRef.current.speak(reply, voicePer);
    } else {
      setState("idle");
    }
  }, [addMessage, getContext, voiceMode, voicePer]);

  // 初始化语音引擎
  const initEngine = useCallback(() => {
    if (engineRef.current) return;
    if (!speechSupported) return;

    engineRef.current = createSpeechEngine({
      onResult: (text) => {
        handleUserInput(text);
      },
      onError: () => {
        setState("idle");
      },
      onStateChange: (s) => {
        if (s === "listening") setState("listening");
        else if (s === "speaking") setState("speaking");
        else if (s === "idle" && state !== "thinking") {
          // 只有不在 thinking 状态时才回到 idle
          // speaking → idle 由 TTS onend 触发
        }
      },
    });
  }, [speechSupported, handleUserInput, state]);

  // 申请麦克风权限
  const requestMic = useCallback(async (): Promise<boolean> => {
    if (micGranted) return true;
    try {
      await navigator.mediaDevices.getUserMedia({ audio: true });
      setMicGranted(true);
      return true;
    } catch {
      setMicGranted(false);
      return false;
    }
  }, [micGranted]);

  // 开始语音输入
  const startVoice = useCallback(async () => {
    if (!speechSupported) return;
    initEngine();
    const ok = await requestMic();
    if (!ok) return;
    engineRef.current?.startListening();
  }, [speechSupported, initEngine, requestMic]);

  // 停止语音输入（用户取消）
  const stopVoice = useCallback(() => {
    engineRef.current?.stopListening();
    if (state === "listening") setState("idle");
  }, [state]);

  // 打断数字人说话
  const interrupt = useCallback(() => {
    engineRef.current?.stopSpeaking();
    setState("listening");
    engineRef.current?.startListening();
  }, []);

  // 清理
  const cleanup = useCallback(() => {
    engineRef.current?.destroy();
    engineRef.current = null;
    setState("idle");
  }, []);

  // 切换语音模式
  const toggleVoiceMode = useCallback(() => setVoiceMode(v => !v), []);

  // 语音模式开启时初始化引擎，关闭时停止播放
  useEffect(() => {
    if (!speechSupported) return;
    if (voiceMode && !engineRef.current) {
      initEngine();
    } else if (!voiceMode) {
      engineRef.current?.stopSpeaking();
    }
  }, [voiceMode]);

  // 从全局同步语音模式（"语音助手"按钮已开启时，面板内小喇叭同步亮起）
  useEffect(() => {
    if (store.voiceMode) setVoiceMode(true);
  }, [store.voiceMode]);

  // 标记首次引导完成
  const dismissFirstTime = useCallback(() => {
    markFirstTimeDone();
    setShowWelcome(false);
    // 首次打开数字人主动说话
    const greeting =
      "你好呀！我是花小猪助手，有什么需要帮忙的吗？你可以点击下面的话筒跟我说话，或者试试快捷指令。";
    addMessage({ role: "assistant", content: greeting });
  }, [addMessage]);

  // 组件卸载时清理
  useEffect(() => {
    return () => {
      engineRef.current?.destroy();
    };
  }, []);

  return {
    state,
    messages,
    showWelcome,
    volumeLevel,
    speechSupported,
    micGranted,
    isFirstTime: isFirstTime(),
    addMessage,
    startVoice,
    stopVoice,
    interrupt,
    handleUserInput,
    voiceMode,
    toggleVoiceMode,
    voicePer,
    setVoicePer,
    cleanup,
    dismissFirstTime,
    setState,
  };
}
