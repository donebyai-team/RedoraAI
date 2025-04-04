export const generateColor = (str: string) => {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash)
  }

  const hue = Math.abs(hash) % 360

  const randomSeed = (hash * 9301 + 49297) % 233280
  const random = randomSeed / 233280.0

  const saturation = 40 + Math.floor(random * 60)

  const lightness = 50 + Math.floor(random * 50)

  return `hsl(${hue}, ${saturation}%, ${lightness}%)`
}
