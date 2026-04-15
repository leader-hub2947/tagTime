<template>
  <div class="animated-characters" :style="{ width: '550px', height: '400px' }">
    <!-- Purple tall rectangle character - Back layer -->
    <div
      ref="purpleRef"
      class="character purple-character"
      :style="purpleStyles"
    >
      <!-- Eyes -->
      <div class="eyes-container" :style="purpleEyesStyles">
        <EyeBall
          :size="18"
          :pupilSize="7"
          :maxDistance="5"
          eyeColor="white"
          pupilColor="#2D2D2D"
          :isBlinking="isPurpleBlinking"
          :forceLookX="purpleForceLookX"
          :forceLookY="purpleForceLookY"
        />
        <EyeBall
          :size="18"
          :pupilSize="7"
          :maxDistance="5"
          eyeColor="white"
          pupilColor="#2D2D2D"
          :isBlinking="isPurpleBlinking"
          :forceLookX="purpleForceLookX"
          :forceLookY="purpleForceLookY"
        />
      </div>
    </div>

    <!-- Black tall rectangle character - Middle layer -->
    <div
      ref="blackRef"
      class="character black-character"
      :style="blackStyles"
    >
      <!-- Eyes -->
      <div class="eyes-container" :style="blackEyesStyles">
        <EyeBall
          :size="16"
          :pupilSize="6"
          :maxDistance="4"
          eyeColor="white"
          pupilColor="#2D2D2D"
          :isBlinking="isBlackBlinking"
          :forceLookX="blackForceLookX"
          :forceLookY="blackForceLookY"
        />
        <EyeBall
          :size="16"
          :pupilSize="6"
          :maxDistance="4"
          eyeColor="white"
          pupilColor="#2D2D2D"
          :isBlinking="isBlackBlinking"
          :forceLookX="blackForceLookX"
          :forceLookY="blackForceLookY"
        />
      </div>
    </div>

    <!-- Orange semi-circle character - Front left -->
    <div
      ref="orangeRef"
      class="character orange-character"
      :style="orangeStyles"
    >
      <!-- Eyes - just pupils, no white -->
      <div class="eyes-container" :style="orangeEyesStyles">
        <Pupil
          :size="12"
          :maxDistance="5"
          pupilColor="#2D2D2D"
          :forceLookX="isHidingPassword ? -5 : undefined"
          :forceLookY="isHidingPassword ? -4 : undefined"
        />
        <Pupil
          :size="12"
          :maxDistance="5"
          pupilColor="#2D2D2D"
          :forceLookX="isHidingPassword ? -5 : undefined"
          :forceLookY="isHidingPassword ? -4 : undefined"
        />
      </div>
    </div>

    <!-- Yellow tall rectangle character - Front right -->
    <div
      ref="yellowRef"
      class="character yellow-character"
      :style="yellowStyles"
    >
      <!-- Eyes - just pupils, no white -->
      <div class="eyes-container" :style="yellowEyesStyles">
        <Pupil
          :size="12"
          :maxDistance="5"
          pupilColor="#2D2D2D"
          :forceLookX="isHidingPassword ? -5 : undefined"
          :forceLookY="isHidingPassword ? -4 : undefined"
        />
        <Pupil
          :size="12"
          :maxDistance="5"
          pupilColor="#2D2D2D"
          :forceLookX="isHidingPassword ? -5 : undefined"
          :forceLookY="isHidingPassword ? -4 : undefined"
        />
      </div>
      <!-- Horizontal line for mouth -->
      <div class="mouth" :style="yellowMouthStyles"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import EyeBall from './EyeBall.vue'
import Pupil from './Pupil.vue'

interface Props {
  isTyping?: boolean
  showPassword?: boolean
  passwordLength?: number
}

const props = withDefaults(defineProps<Props>(), {
  isTyping: false,
  showPassword: false,
  passwordLength: 0
})

// Refs for character elements
const purpleRef = ref<HTMLDivElement | null>(null)
const blackRef = ref<HTMLDivElement | null>(null)
const yellowRef = ref<HTMLDivElement | null>(null)
const orangeRef = ref<HTMLDivElement | null>(null)

// Mouse position
const mouseX = ref(0)
const mouseY = ref(0)

// Blinking states
const isPurpleBlinking = ref(false)
const isBlackBlinking = ref(false)

// Animation states
const isLookingAtEachOther = ref(false)
const isPurplePeeking = ref(false)

// Track mouse movement
const handleMouseMove = (e: MouseEvent) => {
  mouseX.value = e.clientX
  mouseY.value = e.clientY
}

onMounted(() => {
  window.addEventListener('mousemove', handleMouseMove)
  startBlinking()
})

onUnmounted(() => {
  window.removeEventListener('mousemove', handleMouseMove)
})

// Blinking effect for purple character
let purpleBlinkTimeout: ReturnType<typeof setTimeout> | null = null
const schedulePurpleBlink = () => {
  const interval = Math.random() * 4000 + 3000
  purpleBlinkTimeout = setTimeout(() => {
    isPurpleBlinking.value = true
    setTimeout(() => {
      isPurpleBlinking.value = false
      schedulePurpleBlink()
    }, 150)
  }, interval)
}

// Blinking effect for black character
let blackBlinkTimeout: ReturnType<typeof setTimeout> | null = null
const scheduleBlackBlink = () => {
  const interval = Math.random() * 4000 + 3000
  blackBlinkTimeout = setTimeout(() => {
    isBlackBlinking.value = true
    setTimeout(() => {
      isBlackBlinking.value = false
      scheduleBlackBlink()
    }, 150)
  }, interval)
}

const startBlinking = () => {
  schedulePurpleBlink()
  scheduleBlackBlink()
}

// Watch for typing state
watch(() => props.isTyping, (newValue) => {
  if (newValue) {
    isLookingAtEachOther.value = true
    setTimeout(() => {
      isLookingAtEachOther.value = false
    }, 800)
  } else {
    isLookingAtEachOther.value = false
  }
})

// Watch for password visibility
let peekTimeout: ReturnType<typeof setTimeout> | null = null
watch(() => [props.passwordLength, props.showPassword], ([length, show]) => {
  const pwdLength = length as number
  const isVisible = show as boolean
  if (pwdLength > 0 && isVisible) {
    const schedulePeek = () => {
      peekTimeout = setTimeout(() => {
        isPurplePeeking.value = true
        setTimeout(() => {
          isPurplePeeking.value = false
        }, 800)
      }, Math.random() * 3000 + 2000)
    }
    schedulePeek()
  } else {
    isPurplePeeking.value = false
    if (peekTimeout) {
      clearTimeout(peekTimeout)
    }
  }
})

// Calculate position for a character
const calculatePosition = (element: HTMLDivElement | null) => {
  if (!element) return { faceX: 0, faceY: 0, bodySkew: 0 }

  const rect = element.getBoundingClientRect()
  const centerX = rect.left + rect.width / 2
  const centerY = rect.top + rect.height / 3

  const deltaX = mouseX.value - centerX
  const deltaY = mouseY.value - centerY

  const faceX = Math.max(-15, Math.min(15, deltaX / 20))
  const faceY = Math.max(-10, Math.min(10, deltaY / 30))
  const bodySkew = Math.max(-6, Math.min(6, -deltaX / 120))

  return { faceX, faceY, bodySkew }
}

// Computed positions
const purplePos = computed(() => calculatePosition(purpleRef.value))
const blackPos = computed(() => calculatePosition(blackRef.value))
const yellowPos = computed(() => calculatePosition(yellowRef.value))
const orangePos = computed(() => calculatePosition(orangeRef.value))

// Computed values
const isHidingPassword = computed(() => props.passwordLength > 0 && !props.showPassword)

const purpleForceLookX = computed(() => {
  if (props.passwordLength && props.showPassword) {
    return isPurplePeeking.value ? 4 : -4
  }
  if (isLookingAtEachOther.value) return 3
  return undefined
})

const purpleForceLookY = computed(() => {
  if (props.passwordLength && props.showPassword) {
    return isPurplePeeking.value ? 5 : -4
  }
  if (isLookingAtEachOther.value) return 4
  return undefined
})

const blackForceLookX = computed(() => {
  if (props.passwordLength && props.showPassword) return -4
  if (isLookingAtEachOther.value) return 0
  return undefined
})

const blackForceLookY = computed(() => {
  if (props.passwordLength && props.showPassword) return -4
  if (isLookingAtEachOther.value) return -4
  return undefined
})

// Styles
const purpleStyles = computed(() => {
  const isHiding = isHidingPassword.value
  const isTyping = props.isTyping
  const pos = purplePos.value

  let transform = ''
  if (props.passwordLength && props.showPassword) {
    transform = 'skewX(0deg)'
  } else if (isTyping || isHiding) {
    transform = `skewX(${(pos.bodySkew || 0) - 12}deg) translateX(40px)`
  } else {
    transform = `skewX(${pos.bodySkew || 0}deg)`
  }

  return {
    left: '70px',
    width: '180px',
    height: (isTyping || isHiding) ? '440px' : '400px',
    backgroundColor: '#6C3FF5',
    zIndex: 1,
    transform,
    transformOrigin: 'bottom center'
  }
})

const purpleEyesStyles = computed(() => {
  const pos = purplePos.value
  const isHiding = props.passwordLength && props.showPassword

  return {
    left: isHiding ? '20px' : isLookingAtEachOther.value ? '55px' : `${45 + pos.faceX}px`,
    top: isHiding ? '35px' : isLookingAtEachOther.value ? '65px' : `${40 + pos.faceY}px`
  }
})

const blackStyles = computed(() => {
  const isHiding = props.passwordLength && props.showPassword
  const pos = blackPos.value

  let transform = ''
  if (isHiding) {
    transform = 'skewX(0deg)'
  } else if (isLookingAtEachOther.value) {
    transform = `skewX(${(pos.bodySkew || 0) * 1.5 + 10}deg) translateX(20px)`
  } else if (props.isTyping || isHidingPassword.value) {
    transform = `skewX(${(pos.bodySkew || 0) * 1.5}deg)`
  } else {
    transform = `skewX(${pos.bodySkew || 0}deg)`
  }

  return {
    left: '240px',
    width: '120px',
    height: '310px',
    backgroundColor: '#2D2D2D',
    zIndex: 2,
    transform,
    transformOrigin: 'bottom center'
  }
})

const blackEyesStyles = computed(() => {
  const pos = blackPos.value
  const isHiding = props.passwordLength && props.showPassword

  return {
    left: isHiding ? '10px' : isLookingAtEachOther.value ? '32px' : `${26 + pos.faceX}px`,
    top: isHiding ? '28px' : isLookingAtEachOther.value ? '12px' : `${32 + pos.faceY}px`
  }
})

const orangeStyles = computed(() => {
  const pos = orangePos.value
  const isHiding = props.passwordLength && props.showPassword

  return {
    left: '0px',
    width: '240px',
    height: '200px',
    zIndex: 3,
    backgroundColor: '#FF9B6B',
    borderRadius: '120px 120px 0 0',
    transform: isHiding ? 'skewX(0deg)' : `skewX(${pos.bodySkew || 0}deg)`,
    transformOrigin: 'bottom center'
  }
})

const orangeEyesStyles = computed(() => {
  const pos = orangePos.value
  const isHiding = props.passwordLength && props.showPassword

  return {
    left: isHiding ? '50px' : `${82 + (pos.faceX || 0)}px`,
    top: isHiding ? '85px' : `${90 + (pos.faceY || 0)}px`
  }
})

const yellowStyles = computed(() => {
  const pos = yellowPos.value
  const isHiding = props.passwordLength && props.showPassword

  return {
    left: '310px',
    width: '140px',
    height: '230px',
    backgroundColor: '#E8D754',
    borderRadius: '70px 70px 0 0',
    zIndex: 4,
    transform: isHiding ? 'skewX(0deg)' : `skewX(${pos.bodySkew || 0}deg)`,
    transformOrigin: 'bottom center'
  }
})

const yellowEyesStyles = computed(() => {
  const pos = yellowPos.value
  const isHiding = props.passwordLength && props.showPassword

  return {
    left: isHiding ? '20px' : `${52 + (pos.faceX || 0)}px`,
    top: isHiding ? '35px' : `${40 + (pos.faceY || 0)}px`
  }
})

const yellowMouthStyles = computed(() => {
  const pos = yellowPos.value
  const isHiding = props.passwordLength && props.showPassword

  return {
    left: isHiding ? '10px' : `${40 + (pos.faceX || 0)}px`,
    top: isHiding ? '88px' : `${88 + (pos.faceY || 0)}px`
  }
})
</script>

<style scoped>
.animated-characters {
  position: relative;
}

.character {
  position: absolute;
  bottom: 0;
  transition: all 0.7s ease-in-out;
}

.purple-character {
  border-radius: 10px 10px 0 0;
}

.black-character {
  border-radius: 8px 8px 0 0;
}

.yellow-character {
  border-radius: 70px 70px 0 0;
}

.eyes-container {
  position: absolute;
  display: flex;
  gap: 32px;
  transition: all 0.7s ease-in-out;
}

.orange-character .eyes-container {
  gap: 32px;
  transition: all 0.2s ease-out;
}

.yellow-character .eyes-container {
  gap: 24px;
  transition: all 0.2s ease-out;
}

.mouth {
  position: absolute;
  width: 80px;
  height: 4px;
  background: #2D2D2D;
  border-radius: 2px;
  transition: all 0.2s ease-out;
}
</style>
