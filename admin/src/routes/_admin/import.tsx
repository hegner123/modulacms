import { useState, useRef } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import type { ImportFormat, ImportResponse } from '@modulacms/admin-sdk'
import { Upload, FileUp, CheckCircle, XCircle } from 'lucide-react'
import { sdk } from '@/lib/sdk'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

export const Route = createFileRoute('/_admin/import')({
  component: ImportPage,
})

const FORMATS: { value: ImportFormat; label: string }[] = [
  { value: 'contentful', label: 'Contentful' },
  { value: 'sanity', label: 'Sanity' },
  { value: 'strapi', label: 'Strapi' },
  { value: 'wordpress', label: 'WordPress' },
  { value: 'clean', label: 'Clean' },
]

function ImportPage() {
  const [format, setFormat] = useState<ImportFormat>('contentful')
  const [jsonInput, setJsonInput] = useState('')
  const [isImporting, setIsImporting] = useState(false)
  const [result, setResult] = useState<ImportResponse | null>(null)
  const [importError, setImportError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  function handleFileUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (event) => {
      const text = event.target?.result
      if (typeof text === 'string') {
        setJsonInput(text)
      }
    }
    reader.readAsText(file)

    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  async function handleImport() {
    if (!jsonInput.trim()) return

    setIsImporting(true)
    setResult(null)
    setImportError(null)

    try {
      const data = JSON.parse(jsonInput) as Record<string, unknown>
      const response = await sdk.import[format](data)
      setResult(response)
    } catch (err) {
      if (err instanceof SyntaxError) {
        setImportError('Invalid JSON: ' + err.message)
      } else {
        setImportError(err instanceof Error ? err.message : 'Import failed')
      }
    } finally {
      setIsImporting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Import</h1>
        <p className="text-muted-foreground">
          Import content, schemas, and configurations from external sources.
        </p>
      </div>

      <Tabs
        value={format}
        onValueChange={(v) => setFormat(v as ImportFormat)}
      >
        <TabsList>
          {FORMATS.map((f) => (
            <TabsTrigger key={f.value} value={f.value}>
              {f.label}
            </TabsTrigger>
          ))}
        </TabsList>

        {FORMATS.map((f) => (
          <TabsContent key={f.value} value={f.value}>
            <Card>
              <CardHeader>
                <CardTitle>Import from {f.label}</CardTitle>
                <CardDescription>
                  Paste your {f.label} export JSON below or upload a JSON file.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center gap-4">
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept=".json,application/json"
                    className="hidden"
                    onChange={handleFileUpload}
                  />
                  <Button
                    variant="outline"
                    onClick={() => fileInputRef.current?.click()}
                  >
                    <FileUp className="mr-2 h-4 w-4" />
                    Upload JSON File
                  </Button>
                  {jsonInput.trim() && (
                    <span className="text-sm text-muted-foreground">
                      {jsonInput.length.toLocaleString()} characters loaded
                    </span>
                  )}
                </div>

                <Textarea
                  placeholder={'Paste your ' + f.label + ' export JSON here...'}
                  value={jsonInput}
                  onChange={(e) => setJsonInput(e.target.value)}
                  rows={12}
                  className="font-mono text-sm"
                />

                <Button
                  onClick={handleImport}
                  disabled={!jsonInput.trim() || isImporting}
                >
                  <Upload className="mr-2 h-4 w-4" />
                  {isImporting ? 'Importing...' : 'Import'}
                </Button>
              </CardContent>
            </Card>
          </TabsContent>
        ))}
      </Tabs>

      {importError && (
        <Card className="border-destructive">
          <CardHeader>
            <div className="flex items-center gap-2">
              <XCircle className="h-5 w-5 text-destructive" />
              <CardTitle className="text-destructive">Import Failed</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <p className="text-sm">{importError}</p>
          </CardContent>
        </Card>
      )}

      {result && (
        <Card className={result.success ? 'border-green-500' : 'border-destructive'}>
          <CardHeader>
            <div className="flex items-center gap-2">
              {result.success ? (
                <CheckCircle className="h-5 w-5 text-green-500" />
              ) : (
                <XCircle className="h-5 w-5 text-destructive" />
              )}
              <CardTitle>
                {result.success ? 'Import Successful' : 'Import Completed with Errors'}
              </CardTitle>
            </div>
            {result.message && (
              <CardDescription>{result.message}</CardDescription>
            )}
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-3 gap-4">
              <div className="rounded-lg border p-3 text-center">
                <div className="text-2xl font-bold">{result.datatypes_created}</div>
                <div className="text-sm text-muted-foreground">Datatypes Created</div>
              </div>
              <div className="rounded-lg border p-3 text-center">
                <div className="text-2xl font-bold">{result.fields_created}</div>
                <div className="text-sm text-muted-foreground">Fields Created</div>
              </div>
              <div className="rounded-lg border p-3 text-center">
                <div className="text-2xl font-bold">{result.content_created}</div>
                <div className="text-sm text-muted-foreground">Content Created</div>
              </div>
            </div>

            {result.errors.length > 0 && (
              <div className="space-y-2">
                <p className="text-sm font-medium">
                  Errors ({result.errors.length}):
                </p>
                <div className="max-h-48 space-y-1 overflow-y-auto rounded-md border p-3">
                  {result.errors.map((error, index) => (
                    <p key={index} className="text-sm text-destructive">
                      {error}
                    </p>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
