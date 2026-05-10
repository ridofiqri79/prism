import { onBeforeUnmount, ref } from 'vue'
import type { TreeNode } from 'primevue/treenode'
import { usePagination, type SortOrder } from '@/composables/usePagination'
import type { PaginationMeta } from '@/types/api.types'
import type { ListParams } from '@/types/master.types'

// Re-exported from shared utils for backward compatibility
export { toFormErrors, type FormErrors } from '@/utils/form-errors'

export interface HierarchyItem {
  id: string
  parent_id?: string
  has_children?: boolean
}

export interface CodeHierarchyItem {
  id: string
  code: string
  parent_code?: string
  has_children?: boolean
}

export interface AppTreeNode<T> extends Omit<TreeNode, 'data' | 'children'> {
  key: string
  data: T
  children?: AppTreeNode<T>[]
}

export function buildIdTree<T extends HierarchyItem>(items: T[]): AppTreeNode<T>[] {
  const nodes = new Map<string, AppTreeNode<T>>()

  for (const item of items) {
    nodes.set(item.id, { key: item.id, data: item, children: [] })
  }

  const roots: AppTreeNode<T>[] = []

  for (const item of items) {
    const node = nodes.get(item.id)
    if (!node) continue

    if (item.parent_id && nodes.has(item.parent_id)) {
      nodes.get(item.parent_id)?.children?.push(node)
    } else {
      roots.push(node)
    }
  }

  return roots
}

export function buildCodeTree<T extends CodeHierarchyItem>(items: T[]): AppTreeNode<T>[] {
  const nodes = new Map<string, AppTreeNode<T>>()

  for (const item of items) {
    nodes.set(item.code, { key: item.id, data: item, children: [] })
  }

  const roots: AppTreeNode<T>[] = []

  for (const item of items) {
    const node = nodes.get(item.code)
    if (!node) continue

    if (item.parent_code && nodes.has(item.parent_code)) {
      nodes.get(item.parent_code)?.children?.push(node)
    } else {
      roots.push(node)
    }
  }

  return roots
}

export function buildLazyIdNodes<T extends HierarchyItem>(items: T[]): AppTreeNode<T>[] {
  return items.map((item) => ({
    key: item.id,
    data: item,
    leaf: !item.has_children,
  }))
}

export function buildLazyCodeNodes<T extends CodeHierarchyItem>(items: T[]): AppTreeNode<T>[] {
  return items.map((item) => ({
    key: item.id,
    data: item,
    leaf: !item.has_children,
  }))
}

export function useMasterListControls(defaultSort: string, defaultOrder: SortOrder = 'asc') {
  const pagination = usePagination()
  pagination.sort.value = defaultSort
  pagination.order.value = defaultOrder

  const total = ref(0)
  const loading = ref(false)
  const search = ref('')
  let searchTimer: ReturnType<typeof setTimeout> | undefined

  function params(extra: ListParams = {}): ListParams {
    return {
      ...pagination.queryParams.value,
      search: search.value.trim() || undefined,
      ...extra,
    }
  }

  function syncMeta(meta: PaginationMeta) {
    total.value = meta.total
    pagination.page.value = meta.page
    pagination.limit.value = meta.limit
  }

  function handlePage(value: number, loadData: () => void | Promise<void>) {
    pagination.page.value = value
    void loadData()
  }

  function handleLimit(value: number, loadData: () => void | Promise<void>) {
    pagination.limit.value = value
    pagination.resetPage()
    void loadData()
  }

  function handleSort(value: { sort: string; order: SortOrder }, loadData: () => void | Promise<void>) {
    pagination.sort.value = value.sort
    pagination.order.value = value.order
    pagination.resetPage()
    void loadData()
  }

  function resetAndLoad(loadData: () => void | Promise<void>) {
    pagination.resetPage()
    void loadData()
  }

  function resetAndLoadDebounced(loadData: () => void | Promise<void>) {
    pagination.resetPage()
    if (searchTimer) {
      clearTimeout(searchTimer)
    }
    searchTimer = setTimeout(() => {
      void loadData()
    }, 250)
  }

  onBeforeUnmount(() => {
    if (searchTimer) {
      clearTimeout(searchTimer)
    }
  })

  return {
    pagination,
    total,
    loading,
    search,
    params,
    syncMeta,
    handlePage,
    handleLimit,
    handleSort,
    resetAndLoad,
    resetAndLoadDebounced,
  }
}
