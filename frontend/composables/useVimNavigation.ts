import { ref, type Ref } from 'vue'

export interface UseVimNavigationOptions {
  onSelectionChange?: (index: number | null) => void
}

export interface VimNavigationReturn {
  selectedIndex: Ref<number | null>
  isFocused: Ref<boolean>
  setItemCount: (count: number) => void
  setFocused: (focused: boolean) => void
  moveDown: () => void
  moveUp: () => void
  clearSelection: () => void
  getSelectedIndex: () => number | null
}

export function createVimNavigation(
  initialItemCount: number,
  options: UseVimNavigationOptions = {}
): VimNavigationReturn {
  const selectedIndex = ref<number | null>(null)
  const isFocused = ref(false)
  let itemCount = initialItemCount

  const notifySelectionChange = () => {
    options.onSelectionChange?.(selectedIndex.value)
  }

  const setItemCount = (_count: number) => {
    itemCount = _count
    if (selectedIndex.value !== null && _count === 0) {
      selectedIndex.value = null
    } else if (selectedIndex.value !== null && selectedIndex.value >= _count) {
      selectedIndex.value = _count > 0 ? _count - 1 : null
    }
  }

  const setFocused = (focused: boolean) => {
    isFocused.value = focused
  }

  const moveDown = () => {
    if (!isFocused.value || itemCount === 0) return

    if (selectedIndex.value === null) {
      selectedIndex.value = 0
    } else if (selectedIndex.value < itemCount - 1) {
      selectedIndex.value++
    }

    notifySelectionChange()
  }

  const moveUp = () => {
    if (!isFocused.value || itemCount === 0) return

    if (selectedIndex.value === null) {
      selectedIndex.value = itemCount - 1
    } else if (selectedIndex.value > 0) {
      selectedIndex.value--
    }

    notifySelectionChange()
  }

  const clearSelection = () => {
    selectedIndex.value = null
    notifySelectionChange()
  }

  const getSelectedIndex = () => selectedIndex.value

  return {
    selectedIndex,
    isFocused,
    setItemCount,
    setFocused,
    moveDown,
    moveUp,
    clearSelection,
    getSelectedIndex,
  }
}
