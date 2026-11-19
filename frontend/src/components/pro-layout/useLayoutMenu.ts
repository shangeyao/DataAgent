import { computed, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'
import type { MenuOption } from 'naive-ui'

export type MenuKey = string | number
type MenuMeta = { parentKey?: MenuKey; childrenKeys: MenuKey[] }

interface LayoutMenuOptions {
  menus: MaybeRefOrGetter<MenuOption[]>
  mode?: MaybeRefOrGetter<string>
  accordion?: MaybeRefOrGetter<boolean>
  childrenField?: string
}

interface LayoutMenuReturn {
  horizontalMenuProps: Record<string, never>
  verticalExtraMenuProps: Record<string, never>
  verticalMenuProps: Record<string, unknown>
}

// 内置菜单状态管理，只实现本项目使用的竖直菜单模式。
export function useLayoutMenu(options: LayoutMenuOptions) {
  const activeKey = ref<MenuKey | null>(null)
  const expandedKeys = ref<MenuKey[]>([])
  const childrenField = options.childrenField ?? 'children'
  const menus = computed(() => toValue(options.menus))
  const meta = computed(() => {
    const map = new Map<MenuKey, MenuMeta>()
    const visit = (items: MenuOption[], parentKey?: MenuKey) => {
      for (const item of items) {
        if (item.key === undefined || item.key === null) continue
        const key = item.key
        map.set(key, { parentKey, childrenKeys: [] })
        if (parentKey !== undefined) map.get(parentKey)?.childrenKeys.push(key)
        const children = item[childrenField] as MenuOption[] | undefined
        if (children?.length) visit(children, key)
      }
    }
    visit(menus.value)
    return map
  })
  const fullKeys = computed(() => Array.from(meta.value.keys()))
  const getAncestorKeys = (key: MenuKey | null) => {
    const keys: MenuKey[] = []
    let parent = key === null ? undefined : meta.value.get(key)?.parentKey
    while (parent !== undefined) {
      keys.unshift(parent)
      parent = meta.value.get(parent)?.parentKey
    }
    return keys
  }
  const getMenuKeyFullPath = (key: MenuKey) => {
    const keys: MenuKey[] = [...getAncestorKeys(key), key]
    let children = meta.value.get(key)?.childrenKeys ?? []
    while (children.length) {
      const child = children[0]!
      keys.push(child)
      children = meta.value.get(child)?.childrenKeys ?? []
    }
    return keys
  }
  watch(activeKey, () => {
    const accordion = toValue(options.accordion ?? false)
    const keys = accordion ? getAncestorKeys(activeKey.value) : [...expandedKeys.value, ...getAncestorKeys(activeKey.value)]
    expandedKeys.value = Array.from(new Set(keys))
  })
  const verticalMenuProps = computed(() => ({
    mode: 'vertical' as const,
    options: menus.value,
    value: activeKey.value,
    expandedKeys: expandedKeys.value,
    onUpdateValue: (key: MenuKey) => { activeKey.value = key },
    onUpdateExpandedKeys: (keys: MenuKey[]) => { expandedKeys.value = keys },
  }))
  const layout = computed<LayoutMenuReturn>(() => ({
    horizontalMenuProps: {},
    verticalExtraMenuProps: {},
    verticalMenuProps: verticalMenuProps.value,
  }))
  return { fullKeys, activeKey, getMenuKeyFullPath, layout }
}
