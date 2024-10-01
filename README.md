# advisor
- Работа с терминалом Quik
- Скачивание исторических баров с mfd и finam
- Тестирование на истории торговых стратегий

## Работа с терминалом Quik
Работа с торговым терминалом Quik на языке golang реализована по аналогии с проектом [QuikSharp](https://github.com/finsight/QUIKSharp). Для работы необходимы [Lua скрипты для Quik](https://github.com/finsight/QUIKSharp/tree/master/src/QuikSharp/lua).
```
	var port = 34130

	mainConn, err := quik.InitConnection(port)
	if err != nil {
		return err
	}
	defer mainConn.Close()

	var quikService = quik.NewQuikService(mainConn)

	callbackConn, err := quik.InitConnection(port + 1)
	if err != nil {
		return err
	}
	defer callbackConn.Close()

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		for _, _ = range quik.QuikCallbacks(callbackConn) {
		}
		return nil
	})

	g.Go(func() error {
		defer callbackConn.Close()
		quikService.MessageInfo("Привет из go")
		return nil
	})

	return g.Wait()
```
## Скачивание исторических баров
```
$go run ./cmd/tests update -help
Usage:
  -maxdays int
         (default 30)
  -provider string
    
  -security string
    
  -start value
    
  -timeframe string
         (default "minutes5")
```
Пример использования:
```
$go run ./cmd/tests update -security CNY-12.24,CNYRUBF,CNYRUBTOM
```

## Тестирование на истории торговых стратегий
```
$go run ./cmd/tests history -help
Usage:
  -advisor string
    
  -lever float
         (default 9)
  -multy
         (default true)
  -security string
         (default "CNY")
  -slippage float
         (default 0.0002)
  -startquarter int
    
  -startyear int
         (default 2024)
  -timeframe string
         (default "minutes5")
```
Пример использования:
```
$go run ./cmd/tests history -security Si -startyear 2009 -lever 0 -advisor main
```
