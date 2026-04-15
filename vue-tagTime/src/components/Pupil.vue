<template>
  <div
    ref="pupilRef"
    class="pupil"
    :style="pupilStyles"
  />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface Props {
  size?: number
  maxDistance?: number
  pupilColor?: string
  forceLookX?: number
  forceLookY?: number
}

const props = withDefaults(defineProps<Props>(), {
  size: 12,
  maxDistance: 5,
  pupilColor: 'black'
})

const pupilRef = ref<HTMLDivElement | null>(null)
const mouseX = ref(0)
const mouseY = ref(0)

const handleMouseMove = (e: MouseEvent) => {
  mouseX.value = e.clientX
  mouseY.value = e.clientY
}

onMounted(() => {
  window.addEventListener('mousemove', handleMouseMove)
})

onUnmounted(() => {
  window.removeEventListener('mousemove', handleMouseMove)
})

const position = computed(() => {
  if (props.forceLookX !== undefined && props.forceLookY !== undefined) {
    return { x: props.forceLookX, y: props.forceLookY }
  }

  if (!pupilRef.value) return { x: 0, y: 0 }

  const rect = pupilRef.value.getBoundingClientRect()
  const pupilCenterX = rect.left + rect.width / 2
  const pupilCenterY = rect.top + rect.height / 2

  const deltaX = mouseX.value - pupilCenterX
  const deltaY = mouseY.value - pupilCenterY
  const distance = Math.min(Math.sqrt(deltaX ** 2 + deltaY ** 2), props.maxDistance)

  const angle = Math.atan2(deltaY, deltaX)
  const x = Math.cos(angle) * distance
  const y = Math.sin(angle) * distance

  return { x, y }
})

const pupilStyles = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
  backgroundColor: props.pupilColor,
  borderRadius: '50%',
  transform: `translate(${position.value.x}px, ${position.value.y}px)`,
  transition: 'transform 0.1s ease-out'
}))
</script>

<style scoped>
.pupil {
  border-radius: 50%;
}
</style>
