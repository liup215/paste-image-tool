package storage

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/22569/paste-image-tool/internal/config"
	"golang.org/x/image/draw"
)

// ImageSaver 图片保存器
type ImageSaver struct {
	config *config.Config
	mu     sync.Mutex
}

// NewImageSaver 创建图片保存器
func NewImageSaver(cfg *config.Config) *ImageSaver {
	return &ImageSaver{
		config: cfg,
	}
}

// SaveImage 保存图片到指定目录（异步执行，立即返回路径）
func (s *ImageSaver) SaveImage(img image.Image) (string, error) {
	// 确保保存目录存在
	if err := os.MkdirAll(s.config.SaveDirectory, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 生成文件名
	filename := s.generateFilename()
	filepath := filepath.Join(s.config.SaveDirectory, filename)

	// 异步保存图片，不阻塞主线程
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		
		// 复制图片数据（避免异步处理时原图被修改）
		imgToSave := img
		
		// 处理图片尺寸
		if s.config.AutoCompress && s.config.MaxDimension > 0 {
			imgToSave = s.resizeImage(imgToSave, s.config.MaxDimension)
		}

		// 创建文件
		file, err := os.Create(filepath)
		if err != nil {
			return
		}
		defer file.Close()

		// 根据格式保存图片
		format := strings.ToLower(s.config.ImageFormat)
		switch format {
		case "jpg", "jpeg":
			quality := s.config.JPEGQuality
			if quality <= 0 || quality > 100 {
				quality = 85
			}
			opts := &jpeg.Options{Quality: quality}
			jpeg.Encode(file, imgToSave, opts)
		case "webp":
			// WebP 编码需要额外处理，这里先使用 PNG 作为回退
			png.Encode(file, imgToSave)
		default: // png
			png.Encode(file, imgToSave)
		}
	}()

	// 立即返回路径，不等待保存完成
	return filepath, nil
}

// generateFilename 生成文件名
func (s *ImageSaver) generateFilename() string {
	now := time.Now()
	template := s.config.FilenameTemplate

	// 替换模板变量
	// 支持两种格式：{date} {time} 或 yyyyMMdd HHmmss
	// Go 的日期格式使用参考日期: 2006-01-02 15:04:05
	replacements := map[string]string{
		"{date}":     now.Format("20060102"),
		"{time}":     now.Format("150405"),
		"{timestamp}": fmt.Sprintf("%d", now.Unix()),
		"yyyyMMdd":   now.Format("20060102"),
		"yyyy":       now.Format("2006"),
		"MM":         now.Format("01"),
		"dd":         now.Format("02"),
		"HH":         now.Format("15"),
		"mm":         now.Format("04"),
		"ss":         now.Format("05"),
		"HHmmss":     now.Format("150405"),
	}

	filename := template
	for key, value := range replacements {
		filename = strings.ReplaceAll(filename, key, value)
	}

	// 确保有正确的扩展名
	format := strings.ToLower(s.config.ImageFormat)
	if format == "jpg" {
		format = "jpeg"
	}

	// 如果文件名没有扩展名，添加默认扩展名
	if !strings.Contains(filename, ".") {
		filename = filename + "." + format
	}

	return filename
}

// resizeImage 调整图片尺寸
func (s *ImageSaver) resizeImage(img image.Image, maxDim int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 如果图片尺寸在限制范围内，不调整
	if width <= maxDim && height <= maxDim {
		return img
	}

	// 计算新的尺寸
	var newWidth, newHeight int
	if width > height {
		newWidth = maxDim
		newHeight = int(float64(height) * float64(maxDim) / float64(width))
	} else {
		newHeight = maxDim
		newWidth = int(float64(width) * float64(maxDim) / float64(height))
	}

	// 创建新图片
	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// 使用 Lanczos 算法进行高质量缩放
	draw.CatmullRom.Scale(newImg, newImg.Bounds(), img, bounds, draw.Over, nil)

	return newImg
}

// GetFormattedPath 根据配置返回格式化后的路径
func (s *ImageSaver) GetFormattedPath(absolutePath string) string {
	format := s.config.PathFormat

	switch format {
	case "relative":
		// 获取保存目录的相对路径部分
		// 用户期望的是相对于当前工作目录的简单路径，如 ./fig/image.png
		relPath := s.getSimpleRelativePath(absolutePath)
		return relPath

	case "markdown":
		// Markdown 格式: ![image](path)
		filename := filepath.Base(absolutePath)
		relPath := s.getSimpleRelativePath(absolutePath)
		return fmt.Sprintf("![%s](%s)", filename, relPath)

	case "html":
		// HTML 格式
		relPath := s.getSimpleRelativePath(absolutePath)
		return fmt.Sprintf(`<img src="%s" alt="image" />`, relPath)

	case "url":
		// URL 格式
		return "file://" + strings.ReplaceAll(absolutePath, "\\", "/")

	default: // absolute
		// 绝对路径也使用正斜杠，便于在代码中使用
		return strings.ReplaceAll(absolutePath, "\\", "/")
	}
}

// getSimpleRelativePath 获取简单的相对路径
// 从保存目录中提取相对路径，格式为 ./dir/file.png
func (s *ImageSaver) getSimpleRelativePath(absolutePath string) string {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		// 如果无法获取工作目录，返回正斜杠格式的绝对路径
		return strings.ReplaceAll(absolutePath, "\\", "/")
	}
	
	// 尝试计算相对路径
	rel, err := filepath.Rel(wd, absolutePath)
	if err != nil {
		// 如果无法计算相对路径，返回正斜杠格式的绝对路径
		return strings.ReplaceAll(absolutePath, "\\", "/")
	}
	
	// 转换为正斜杠
	rel = strings.ReplaceAll(rel, "\\", "/")
	
	// 如果路径不以 ./ 开头且不是以 / 开头，添加 ./
	if !strings.HasPrefix(rel, "./") && !strings.HasPrefix(rel, "/") && !strings.HasPrefix(rel, "../") {
		rel = "./" + rel
	}
	
	return rel
}

// GetImageInfo 获取图片信息
func (s *ImageSaver) GetImageInfo(filepath string) (*ImageInfo, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	stat, _ := file.Stat()

	return &ImageInfo{
		Path:   filepath,
		Format: format,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Size:   stat.Size(),
	}, nil
}

// ImageInfo 图片信息
type ImageInfo struct {
	Path   string
	Format string
	Width  int
	Height int
	Size   int64
}
