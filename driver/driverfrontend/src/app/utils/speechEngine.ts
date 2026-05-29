export type SpeechState = "idle" | "listening" | "speaking" | "thinking";

export interface SpeechCallbacks {
  onResult: (text: string) => void;
  onError: (msg: string) => void;
  onStateChange: (state: SpeechState) => void;
  onVolume?: (level: number) => void;
  onIdle?: () => void;
}

const SpeechRecognitionAPI =
  (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;

export function isSpeechSupported(): boolean {
  return !!SpeechRecognitionAPI;
}

export function createSpeechEngine(cbs: SpeechCallbacks) {
  let recognition: any = null;
  let listeningTimer: ReturnType<typeof setTimeout> | null = null;
  let currentState: SpeechState = "idle";
  let currentAudio: HTMLAudioElement | null = null;
  let continuous = false;
  let volumeInterval: ReturnType<typeof setInterval> | null = null;

  function setState(s: SpeechState) {
    currentState = s;
    cbs.onStateChange(s);
  }

  function startListening() {
    if (!SpeechRecognitionAPI) {
      cbs.onError("当前浏览器不支持语音识别");
      return;
    }
    stopSpeaking();

    // 防重入：清理前一个实例和定时器，防止累积
    if (listeningTimer) {
      clearTimeout(listeningTimer);
      listeningTimer = null;
    }
    if (recognition) {
      try { recognition.abort(); } catch {}
      recognition = null;
    }

    recognition = new SpeechRecognitionAPI();
    recognition.lang = "zh-CN";
    recognition.continuous = false;
    recognition.interimResults = false;
    recognition.maxAlternatives = 1;

    recognition.onstart = () => {
      setState("listening");
    };

    recognition.onresult = (event: any) => {
      const text = event.results[0][0].transcript;
      // 已获得结果，取消静音超时定时器，防止后续误重启
      if (listeningTimer) {
        clearTimeout(listeningTimer);
        listeningTimer = null;
      }
      cbs.onResult(text);
    };

    recognition.onerror = (event: any) => {
      if (event.error === "no-speech") {
        cbs.onError("没有听到你说什么");
      } else if (event.error === "aborted") {
      } else if (event.error === "not-allowed") {
        cbs.onError("请允许麦克风权限");
      } else {
        cbs.onError("语音识别出错，请重试");
      }
    };

    recognition.onend = () => {
      // 连续模式下不在 onend 重启（交给 listeningTimer 或 speak 的 onended），
      // 避免和定时器路径竞争导致重复创建实例
      if (!continuous) {
        setState("idle");
      }
    };

    listeningTimer = setTimeout(() => {
      if (recognition) {
        try { recognition.abort(); } catch {}
        if (continuous) {
          // 10 秒无语音：不自动重启，通知消费端自行处理
          cbs.onIdle?.();
        } else {
          cbs.onError("没有听到你说什么");
        }
      }
    }, 10000);

    try {
      recognition.start();
    } catch {
      cbs.onError("启动语音识别失败");
    }
  }

  function startContinuousListening() {
    continuous = true;
    startListening();
  }

  function stopListening() {
    continuous = false;
    if (listeningTimer) {
      clearTimeout(listeningTimer);
      listeningTimer = null;
    }
    if (recognition) {
      try { recognition.abort(); } catch {}
      recognition = null;
    }
    if (volumeInterval) {
      clearInterval(volumeInterval);
      volumeInterval = null;
    }
    if (currentState === "listening") {
      setState("idle");
    }
  }

  const PER_SPD: Record<string, string> = {
    "3": "6", "4": "5", "5003": "5", "5118": "8",
  };

  function speak(text: string, per = "3") {
    stopSpeaking();

    // 暂停连续监听，防止 TTS 播放时麦克风拾取自身声音形成反馈环路
    const wasContinuous = continuous;
    if (wasContinuous) {
      if (listeningTimer) {
        clearTimeout(listeningTimer);
        listeningTimer = null;
      }
      if (recognition) {
        try { recognition.abort(); } catch {}
        recognition = null;
      }
      if (currentState === "listening") setState("idle");
    }

    const audio = new Audio();
    currentAudio = audio;
    setState("speaking");

    fetch(`/api/v1/ai/tts?text=${encodeURIComponent(text)}&per=${per}&spd=${PER_SPD[per] || "5"}`)
      .then((res) => {
        if (!res.ok) throw new Error("TTS failed");
        return res.blob();
      })
      .then((blob) => {
        if (currentAudio !== audio) return;
        const url = URL.createObjectURL(blob);
        audio.src = url;
        audio.onended = () => {
          URL.revokeObjectURL(url);
          if (currentAudio === audio) currentAudio = null;
          setState("idle");
          if (wasContinuous) {
            setTimeout(() => startListening(), 300);
          }
        };
        audio.play().catch(() => {
          setState("idle");
          if (wasContinuous) {
            setTimeout(() => startListening(), 300);
          }
        });
      })
      .catch(() => {
        setState("idle");
        if (wasContinuous) {
          setTimeout(() => startListening(), 300);
        }
      });
  }

  function stopSpeaking() {
    if (currentAudio) {
      currentAudio.pause();
      currentAudio.src = "";
      currentAudio = null;
    }
    if (currentState === "speaking") {
      setState("idle");
    }
  }

  function destroy() {
    continuous = false;
    stopListening();
    stopSpeaking();
  }

  return {
    startListening,
    startContinuousListening,
    stopListening,
    speak,
    stopSpeaking,
    destroy,
    get state() {
      return currentState;
    },
  };
}

export type SpeechEngine = ReturnType<typeof createSpeechEngine>;
