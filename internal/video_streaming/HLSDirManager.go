package video_streaming

import (
	"bufio"
	"context"
	"fmt"
	"gocv.io/x/gocv"
	"log"
	"os"
	"regexp"
	"sync"
)

type HLSDirManager struct {
	config       *HLSConfig
	shardCounter int

	m3u8Desc  *os.File
	descMutex sync.Mutex

	shardLock sync.Mutex
}

func (manager *HLSDirManager) GetConfig() *HLSConfig {
	return manager.config
}

func NewHLSDirManager(config *HLSConfig) *HLSDirManager {
	// Создайте каталог для HLS
	_ = os.MkdirAll(config.dir, os.ModePerm)

	hlsShardPatt, err := regexp.Compile(`\w+.ts`)
	if err != nil {
		panic(err)
	}

	// Путь к m3u8 файлу
	pth := config.dir + "/" + config.m3u8Name

	// Откройте или создайте файл m3u8
	file, err := os.OpenFile(pth, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to open m3u8 file: %w", err))
	}

	// Если файл пустой, инициализируйте его содержимое
	stat, _ := file.Stat()
	if stat.Size() == 0 {
		writer := bufio.NewWriter(file)
		writePanic := func(str string) {
			_, err := writer.WriteString(str + "\n")
			if err != nil {
				panic(err)
			}
		}
		writePanic("#EXTM3U")
		writePanic("#EXT-X-VERSION:3")
		writePanic("#EXT-X-MEDIA-SEQUENCE:0")
		targetDuration := fmt.Sprintf("#EXT-X-TARGETDURATION:%f", float64(config.frameNumPerShard)/config.fps)
		writePanic(targetDuration)
		if err := writer.Flush(); err != nil {
			panic(err)
		}
	}

	// Подсчитайте количество сегментов
	scanner := bufio.NewScanner(file)
	shardCnt := 0
	for scanner.Scan() {
		line := scanner.Text()
		if hlsShardPatt.MatchString(line) {
			shardCnt++
		}
	}

	// Вернитесь в конец файла для дальнейшей записи
	_, err = file.Seek(0, 2)
	if err != nil {
		panic(err)
	}

	return &HLSDirManager{
		config:       config,
		shardCounter: shardCnt,
		m3u8Desc:     file,
	}
}

func (manager *HLSDirManager) WritePatch(shardName string, videoLen float64) error {
	manager.descMutex.Lock()
	defer manager.descMutex.Unlock()

	writer := bufio.NewWriter(manager.m3u8Desc)

	if _, err := writer.WriteString("#EXTINF:" + fmt.Sprintf("%f\n", videoLen)); err != nil {
		return fmt.Errorf("failed to write EXTINF: %w", err)
	}

	if _, err := writer.WriteString(shardName + "\n"); err != nil {
		return fmt.Errorf("failed to write shard name: %w", err)
	}

	if err := writer.Flush(); err != nil {
		fmt.Printf("Failed to flush writer: %v\n", err)
	}

	return nil
}

func (manager *HLSDirManager) AddBatch(_ context.Context, batch []gocv.Mat) error {
	manager.shardLock.Lock()
	defer manager.shardLock.Unlock()
	newShardName := fmt.Sprintf(`shard%d.ts`, manager.shardCounter)
	log.Printf("Adding %d records\n", len(batch))
	writer, err := gocv.VideoWriterFile(
		manager.config.dir+"/"+newShardName,
		"avc1",
		manager.config.fps,
		batch[0].Cols(),
		batch[0].Rows(),
		true,
	)

	if err != nil {
		log.Printf(`failed to open video writer: %v`, err)
		return err
	}
	defer func(writer *gocv.VideoWriter) {
		_ = writer.Close()
	}(writer)
	for _, im := range batch {
		err := writer.Write(im)
		if err != nil {
			log.Printf(`failed to write video: %v`, err)
			return err
		}
	}
	go func() {
		err := manager.WritePatch(newShardName, float64(len(batch))/manager.config.fps)
		if err != nil {
			log.Fatalf(`failed to write patch: %v`, err)
		}
	}()
	manager.shardCounter++
	return nil
}
