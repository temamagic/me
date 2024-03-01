package main

import (
	"github.com/fsnotify/fsnotify"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func watchRecursive(watcher *fsnotify.Watcher, path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Проверяем, является ли объект файлом или папкой
		if info.Mode().IsDir() {
			// Добавляем папку в вотчер
			err = watcher.Add(path)
			if err != nil {
				log.Println("Ошибка добавления папки в вотчер:", err)
			} else {
				log.Println("Добавлена папка в вотчер:", path)
			}
		}
		return nil
	})
	return err
}

func copyDir(src, dst string) error {
	// Получаем информацию о исходной директории
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Создаем директорию назначения
	err = os.MkdirAll(dst, info.Mode())
	if err != nil {
		return err
	}

	// Получаем содержимое исходной директории
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	// Копируем файлы и рекурсивно копируем поддиректории
	for _, file := range files {
		srcFile := filepath.Join(src, file.Name())
		dstFile := filepath.Join(dst, file.Name())

		if file.IsDir() {
			// Рекурсивно копируем поддиректорию
			err = copyDir(srcFile, dstFile)
			if err != nil {
				return err
			}
		} else {
			// Копируем файл
			err = copyFile(srcFile, dstFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	// Открываем исходный файл для чтения
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Создаем назначенный файл
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Копируем содержимое
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Копирование завершено успешно
	return nil
}
