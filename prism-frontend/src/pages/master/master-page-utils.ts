import type { ZodError } from 'zod'
import type { TreeNode } from 'primevue/treenode'

export type FormErrors<T extends string> = Partial<Record<T, string>>

export function toFormErrors<T extends string>(error: ZodError, fields: readonly T[]) {
  const allowedFields = new Set<string>(fields)
  const errors: FormErrors<T> = {}

  for (const issue of error.issues) {
    const field = String(issue.path[0] ?? '')

    if (allowedFields.has(field) && !errors[field as T]) {
      errors[field as T] = issue.message
    }
  }

  return errors
}

export interface HierarchyItem {
  id: string
  parent_id?: string
}

export interface CodeHierarchyItem {
  id: string
  code: string
  parent_code?: string
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
