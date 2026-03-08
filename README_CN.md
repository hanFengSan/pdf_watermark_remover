# PDF 水印移除工具

这是一个本地工具, 会先将 PDF 页面渲染为图片，再执行水印抑制与可读性修复，最后重建为黑白输出 PDF。当前处理流程针对文字类 PDF 做了特化优化。

## 使用指南
1. 将可执行文件（`.exe`）和待处理 PDF 放在同一目录下。
   - 注意：文件名包含 `remove_watermark` 的 PDF 会被自动跳过。
2. 运行程序（100 页 PDF 约耗时 3 分钟，视机器配置和文档复杂度而定）。
3. 在当前目录查看输出结果。
4. 其他用法：
   - 命令行指定单个 PDF：`pdf_watermark_remover xx/path/xxx.pdf`
   - 输出文件会生成在源 PDF 所在目录。


## 处理流程

1. 校验输入 PDF。
2. 将每页渲染为 PNG（默认 400 DPI）。
3. 估计水印模式并选择页面抑制策略。
4. 执行水印抑制、二值化与可读性清理。
5. 对残影明显页面执行重试策略。
6. 将处理后的图片重建为输出 PDF。

## 构建

要求：

- Go 1.25+

构建当前平台可执行文件：

```bash
go build -o pdf_watermark_remover ./cmd/pdf_watermark_remover
```

构建 Windows amd64 版本：

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/pdf_watermark_remover_windows_amd64.exe ./cmd/pdf_watermark_remover
```
