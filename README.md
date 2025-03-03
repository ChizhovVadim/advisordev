# advisor
- Работа с терминалом Quik
- Автоматическая торговля в Quik по заданным стратегиям
- Скачивание исторических баров с mfd и finam
- Тестирование на истории торговых стратегий

## Команды
- Тестирует на истории торгового советника.
```
$go run ./cmd/history report -help
Usage:
  -advisor string
    
  -finishquarter int
         (default 3)
  -finishyear int
         (default 2025)
  -lever float
    
  -multy
         (default true)
  -security string
    
  -slippage float
         (default 0.0002232)
  -startquarter int
    
  -startyear int
         (default 2025)
  -timeframe string
         (default "minutes5")
```
Пример использования:
```
$go run ./cmd/history report -security Si -startyear 2009 -advisor main
```

- Показывает несколько последних позиций торгового советника (для отладки).
```
$go run ./cmd/history status -security Si-3.25 -advisor main
```

- Запускает автоматическую торговлю.
```
$go run ./cmd/trader
```

- Скачивает исторические котировки (для отладки).
```
$go run ./cmd/history testdownload -security Si-3.25 -provider finam -timeframe minutes5
```

- Скачивает и обновляет исторические котировки.
```
$go run ./cmd/history update --help
Usage:
  -provider string
    
  -security string
    
  -timeframe string
         (default "minutes5")
```
Пример использования:
```
$go run ./cmd/history update -security CNY-12.24,Si-12.24 -provider finam
```
