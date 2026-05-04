type AudioContextConstructor = typeof AudioContext

interface WindowWithWebkitAudio extends Window {
  webkitAudioContext?: AudioContextConstructor
}

function getAudioContextConstructor(): AudioContextConstructor | null {
  if (typeof window === 'undefined') {
    return null
  }

  const audioWindow = window as WindowWithWebkitAudio
  return window.AudioContext ?? audioWindow.webkitAudioContext ?? null
}

export function playMessageTone(): void {
  const AudioContextCtor = getAudioContextConstructor()
  if (!AudioContextCtor) {
    return
  }

  try {
    const audioContext = new AudioContextCtor()
    const oscillator = audioContext.createOscillator()
    const gainNode = audioContext.createGain()

    oscillator.type = 'sine'
    oscillator.frequency.setValueAtTime(880, audioContext.currentTime)
    gainNode.gain.setValueAtTime(0.0001, audioContext.currentTime)
    gainNode.gain.exponentialRampToValueAtTime(0.08, audioContext.currentTime + 0.02)
    gainNode.gain.exponentialRampToValueAtTime(0.0001, audioContext.currentTime + 0.18)

    oscillator.connect(gainNode)
    gainNode.connect(audioContext.destination)

    oscillator.start()
    oscillator.stop(audioContext.currentTime + 0.18)
    oscillator.onended = () => {
      void audioContext.close()
    }
  } catch {
    // Browsers can block audio playback before user interaction; fail quietly.
  }
}
