import { Input, Segmented, Space, Typography, Upload, message } from 'antd'
import type { UploadFile } from 'antd'
import { useEffect, useMemo, useRef } from 'react'
import type { Dispatch, ReactNode, RefObject, SetStateAction } from 'react'
import type { Editor as TiptapEditor } from '@tiptap/core'
import { EditorContent, EditorContext, useEditor } from '@tiptap/react'
import { StarterKit } from '@tiptap/starter-kit'
import { Image } from '@tiptap/extension-image'
import { TaskItem, TaskList } from '@tiptap/extension-list'
import Placeholder from '@tiptap/extension-placeholder'
import { TextAlign } from '@tiptap/extension-text-align'
import { Typography as TiptapTypography } from '@tiptap/extension-typography'
import { Underline } from '@tiptap/extension-underline'
import { Highlight } from '@tiptap/extension-highlight'
import { Subscript } from '@tiptap/extension-subscript'
import { Superscript } from '@tiptap/extension-superscript'
import { Selection } from '@tiptap/extensions'

import { Toolbar, ToolbarGroup, ToolbarSeparator } from '@/components/tiptap-ui-primitive/toolbar'
import { HeadingDropdownMenu } from '@/components/tiptap-ui/heading-dropdown-menu'
import { ImageUploadButton } from '@/components/tiptap-ui/image-upload-button'
import { ListDropdownMenu } from '@/components/tiptap-ui/list-dropdown-menu'
import { BlockquoteButton } from '@/components/tiptap-ui/blockquote-button'
import { CodeBlockButton } from '@/components/tiptap-ui/code-block-button'
import { ColorHighlightPopover } from '@/components/tiptap-ui/color-highlight-popover'
import { LinkPopover } from '@/components/tiptap-ui/link-popover'
import { MarkButton } from '@/components/tiptap-ui/mark-button'
import { TextAlignButton } from '@/components/tiptap-ui/text-align-button'
import { UndoRedoButton } from '@/components/tiptap-ui/undo-redo-button'
import { ImageUploadNode } from '@/components/tiptap-node/image-upload-node/image-upload-node-extension'
import { HorizontalRule } from '@/components/tiptap-node/horizontal-rule-node/horizontal-rule-node-extension'
import { NodeBackground } from '@/components/tiptap-extension/node-background-extension'
import '@/components/tiptap-node/blockquote-node/blockquote-node.scss'
import '@/components/tiptap-node/code-block-node/code-block-node.scss'
import '@/components/tiptap-node/horizontal-rule-node/horizontal-rule-node.scss'
import '@/components/tiptap-node/list-node/list-node.scss'
import '@/components/tiptap-node/image-node/image-node.scss'
import '@/components/tiptap-node/heading-node/heading-node.scss'
import '@/components/tiptap-node/paragraph-node/paragraph-node.scss'

import type { MailAttachment, SMTPTestDraft } from '../types'

type Props = {
  value: SMTPTestDraft
  onChange: Dispatch<SetStateAction<SMTPTestDraft>>
}

type SelectionRange = { from: number; to: number }

const mailImageMaxSize = 3 * 1024 * 1024

function escapeHtml(text: string) {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function toPlainText(html: string) {
  if (!html.trim()) return ''
  const doc = new DOMParser().parseFromString(html, 'text/html')
  return doc.body.textContent || ''
}

function normalizeText(text: string) {
  return text.replace(/\r\n/g, '\n')
}

function plainTextToHtml(text: string) {
  const normalized = normalizeText(text)
  if (!normalized.trim()) return '<p><br></p>'
  return normalized
    .split('\n')
    .map((line) => `<p>${line ? escapeHtml(line) : '<br>'}</p>`)
    .join('')
}

function draftToHtml(value: SMTPTestDraft) {
  return value.html || plainTextToHtml(value.text)
}

function readFileAsDataUrl(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(reader.error || new Error('读取附件失败'))
    reader.readAsDataURL(file)
  })
}

function makeAttachment(file: File, contentBase64: string, inline: boolean): MailAttachment {
  return {
    uid: `${file.name}-${file.size}-${file.lastModified}-${Math.random().toString(36).slice(2, 8)}`,
    filename: file.name,
    content_type: file.type || 'application/octet-stream',
    content_base64: contentBase64,
    content_id: inline ? `img-${Date.now()}-${Math.random().toString(36).slice(2, 8)}` : '',
    inline,
  }
}

function stripDataUrl(dataUrl: string) {
  const idx = dataUrl.indexOf(',')
  return idx >= 0 ? dataUrl.slice(idx + 1) : dataUrl
}

function FieldRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="mail-field">
      <Typography.Text className="mail-field-label">{label}</Typography.Text>
      {children}
    </label>
  )
}

function RichEditor({
  editor,
  toolbarRef,
}: {
  editor: TiptapEditor
  toolbarRef: RefObject<HTMLDivElement | null>
}) {
  return (
    <div className="mail-editor-shell mail-simple-editor-shell">
      <EditorContext.Provider value={{ editor }}>
        <Toolbar ref={toolbarRef} className="mail-simple-toolbar">
          <ToolbarGroup>
            <UndoRedoButton action="undo" />
            <UndoRedoButton action="redo" />
          </ToolbarGroup>
          <ToolbarSeparator />
          <ToolbarGroup>
            <HeadingDropdownMenu modal={false} levels={[1, 2, 3, 4]} />
            <ListDropdownMenu modal={false} types={['bulletList', 'orderedList', 'taskList']} />
            <BlockquoteButton />
            <CodeBlockButton />
          </ToolbarGroup>
          <ToolbarSeparator />
          <ToolbarGroup>
            <MarkButton type="bold" />
            <MarkButton type="italic" />
            <MarkButton type="strike" />
            <MarkButton type="code" />
            <MarkButton type="underline" />
            <ColorHighlightPopover />
            <LinkPopover />
          </ToolbarGroup>
          <ToolbarSeparator />
          <ToolbarGroup>
            <MarkButton type="superscript" />
            <MarkButton type="subscript" />
          </ToolbarGroup>
          <ToolbarSeparator />
          <ToolbarGroup>
            <TextAlignButton align="left" />
            <TextAlignButton align="center" />
            <TextAlignButton align="right" />
            <TextAlignButton align="justify" />
          </ToolbarGroup>
          <ToolbarSeparator />
          <ToolbarGroup>
            <ImageUploadButton />
          </ToolbarGroup>
        </Toolbar>
        <EditorContent editor={editor} role="presentation" className="mail-simple-editor-content" />
      </EditorContext.Provider>
    </div>
  )
}

function PlainEditor({
  value,
  onChange,
}: {
  value: SMTPTestDraft
  onChange: Dispatch<SetStateAction<SMTPTestDraft>>
}) {
  return (
    <div className="mail-editor-shell mail-plain-shell">
      <Input.TextArea
        className="mail-plain-textarea"
        autoSize={{ minRows: 10, maxRows: 16 }}
        value={value.text}
        placeholder="请输入纯文本邮件正文"
        onChange={(e) => {
          const nextText = normalizeText(e.target.value)
          onChange((prev) => ({ ...prev, text: nextText, html: '' }))
        }}
      />
    </div>
  )
}

export function MailComposer({ value, onChange }: Props) {
  const selectionRef = useRef<SelectionRange | null>(null)
  const lastSyncedHtmlRef = useRef<string>(draftToHtml(value))
  const toolbarRef = useRef<HTMLDivElement | null>(null)

  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        horizontalRule: false,
        heading: { levels: [1, 2, 3] },
        link: {
          openOnClick: false,
          enableClickSelection: true,
        },
      }),
      HorizontalRule,
      TextAlign.configure({ types: ['heading', 'paragraph'] }),
      TaskList,
      TaskItem.configure({ nested: true }),
      Underline,
      Highlight.configure({ multicolor: true }),
      Image.configure({
        inline: false,
        allowBase64: true,
      }),
      TiptapTypography,
      Superscript,
      Subscript,
      Selection,
      NodeBackground,
      ImageUploadNode.configure({
        accept: 'image/*',
        limit: 3,
        maxSize: mailImageMaxSize,
        upload: async (file) => {
          const dataUrl = await readFileAsDataUrl(file)
          const attachment = makeAttachment(file, stripDataUrl(dataUrl), true)
          onChange((prev) => ({ ...prev, attachments: [...prev.attachments, attachment] }))
          return dataUrl
        },
        onError: (error) => message.error(error.message || '图片上传失败'),
      }),
      Placeholder.configure({
        placeholder: '编写邮件正文，可插入图片和格式化文本',
      }),
    ],
    content: draftToHtml(value),
    editorProps: {
      attributes: {
        class: 'mail-tiptap-editor',
      },
    },
    onSelectionUpdate: ({ editor }) => {
      selectionRef.current = {
        from: editor.state.selection.from,
        to: editor.state.selection.to,
      }
    },
    onUpdate: ({ editor }) => {
      const html = editor.isEmpty ? '<p><br></p>' : editor.getHTML()
      const text = normalizeText(editor.getText())
      lastSyncedHtmlRef.current = html
      onChange((prev) => ({ ...prev, html, text }))
    },
    immediatelyRender: false,
  })

  useEffect(() => {
    if (!editor || !value.richText) return
    const nextHtml = draftToHtml(value)
    if (value.html !== lastSyncedHtmlRef.current && nextHtml !== lastSyncedHtmlRef.current) {
      editor.commands.setContent(nextHtml, { emitUpdate: false })
      lastSyncedHtmlRef.current = nextHtml
    }
  }, [editor, value.html, value.richText, value.text])

  const attachmentFiles = useMemo(() => {
    return value.attachments.map((item) => ({
      uid: item.uid,
      name: item.filename,
      status: 'done' as const,
      size: 0,
      type: item.content_type,
    }))
  }, [value.attachments])

  const fileList: UploadFile[] = attachmentFiles

  const handleFiles = async (files: File[]) => {
    if (files.length === 0) return
    const inserted: MailAttachment[] = []
    for (const file of files) {
      const dataUrl = await readFileAsDataUrl(file)
      const contentBase64 = stripDataUrl(dataUrl)
      const isImage = file.type.startsWith('image/')
      const attachment = makeAttachment(file, contentBase64, isImage)
      inserted.push(attachment)
      if (value.richText && isImage && editor) {
        const range = selectionRef.current || {
          from: editor.state.selection.from,
          to: editor.state.selection.to,
        }
        editor
          .chain()
          .focus()
          .setTextSelection(range)
          .setImage({ src: dataUrl })
          .run()
      }
    }
    onChange((prev) => ({ ...prev, attachments: [...prev.attachments, ...inserted] }))
    message.success(`已添加 ${files.length} 个附件`)
  }

  const toggleMode = (next: 'rich' | 'plain') => {
    const richText = next === 'rich'
    onChange((prev) => {
      if (richText) {
        const nextHtml = prev.html || (prev.text ? plainTextToHtml(prev.text) : '<p><br></p>')
        return { ...prev, richText: true, html: nextHtml }
      }
      return { ...prev, richText: false, html: '', text: prev.text || toPlainText(prev.html) }
    })
  }

  return (
    <div className="mail-composer">
      <Space direction="vertical" size={12} className="full">
        <div className="mail-meta-grid">
          <FieldRow label="From">
            <Input value={value.from} onChange={(e) => onChange((prev) => ({ ...prev, from: e.target.value }))} placeholder="请输入发件人" />
          </FieldRow>
          <FieldRow label="To">
            <Input value={value.to} onChange={(e) => onChange((prev) => ({ ...prev, to: e.target.value }))} placeholder="多个收件人请用逗号分隔" />
          </FieldRow>
          <FieldRow label="Subject">
            <Input value={value.subject} onChange={(e) => onChange((prev) => ({ ...prev, subject: e.target.value }))} placeholder="请输入主题" />
          </FieldRow>
        </div>
        <div className="mail-mode-switch">
          <Segmented
            value={value.richText ? 'rich' : 'plain'}
            size="small"
            options={[
              { label: '富文本', value: 'rich' },
              { label: '纯文本', value: 'plain' },
            ]}
            onChange={(next) => toggleMode(next as 'rich' | 'plain')}
          />
        </div>
        {value.richText && editor ? (
          <RichEditor editor={editor} toolbarRef={toolbarRef} />
        ) : (
          <PlainEditor value={value} onChange={onChange} />
        )}
        <Space direction="vertical" size={8} className="full">
          <Typography.Text strong>附件</Typography.Text>
          <Upload.Dragger
            multiple
            accept="*"
            beforeUpload={async (file) => {
              await handleFiles([file])
              return false
            }}
            fileList={fileList}
            onRemove={(file) => {
              const next = value.attachments.filter((item) => item.uid !== file.uid)
              onChange((prev) => ({ ...prev, attachments: next }))
              return true
            }}
          >
            <p className="ant-upload-drag-icon">拖拽文件到这里，或点击选择</p>
            <p className="ant-upload-text">支持图片、文档和其他附件</p>
          </Upload.Dragger>
        </Space>
      </Space>
    </div>
  )
}
