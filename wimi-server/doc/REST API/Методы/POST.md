# POST

## DecodeMML

### Описание

| Property | Value |
|----------|-------|
| **Назначение** | Запрос для дешифровки .mml файла. |
| **Метод** | POST |
| **URL** | `/DecodeMML` |

### Структура JSON

#### Поля

| Property | Description |
|----------|-------------|
| **сoded** | Параметр для закодированной модели. |

### Пример запроса

```bash
curl http://127.0.0.1:8081/DecodeMML -H "Content-Type: application/json" -X POST -d '{ "relationType" : "simple" ... }'
```

```JSON
{
  "coded": "7466f60a208d24fe42e23f707a589882f619dc5739d36104a408e8085957ff17aed423252523eddab"
}
```

### Пример ответа

```JSON
{
  "decoded": "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<root>\n <parameters>\n <parametr type=\"double\" id=\"P2015-02-06104032484-02799\" shortName=\"param2\" level=\"root\" classType=\"1\" levelUID=\"0\"/>\n <parametr type=\"double\" id=\"P2015-02-06104025959-00842\" shortName=\"param1\" level=\"root\" classType=\"1\" defaultValue=\"6.00\" levelUID=\"0\"/>\n </parameters>\n <rules>\n <rule id=\"R2015-02-06104133832-00835\" resultId=\"P2015-02-06104032484-02799\" relation=\"O2015-02-06104046321-02145\" initId=\"P2015-02-06104025959-00842\" description=\"test1\"/>\n </rules>\n <relations>\n <relation outObj=\"y\" inputTypes=\"double\" id=\"O2015-02-06104046321-02145\" relationType=\"simple\" outputTypes=\"double\" inObj=\"x\">y=x+2</relation>\n </relations>\n <classes/>\n <constraints/>\n</root>\n"
}
```

## ModelCalc

### Описание

| Property | Value |
|----------|-------|
| **Назначение** | Запрос на вычисление модели. |
| **Метод** | POST |
| **URL** | `/ModelCalc` |

### Структура JSON

#### Поля

| Property | Description |
|----------|-------------|
| **modelID** | ID модели, это любой набор символов. |
| **incommingParameters** | Массив известных параметров, состоит из ID параметра и его значения. |
| **outputParameters** | Массив исковых, состоит только из ID параметров. |
| **outputfFields** | Массив, содержащий ключи формата вывода результата работы API. |
| **service** | Опциональные параметры запроса записываются в тэг. |

#### outputFields (допустимые значения)

| Property | Description |
|----------|-------------|
| **algorithm** | Выводит все содержание тега algorithm. |
| **requiredExploredParameters** | Выводит массив требуемых и найденных параметров. |
| **requiredNotExploredParameters** | Выводит массив требуемых, но не найденных параметров. |
| **notRequiredExploredParameters** | Выводит массив найденных, но не требуемых параметров. |
| **timing** | Выводит время потраченное на разбор запроса, его обработку и формирование ответа в миллисекундах |

> ВНИМАНИЕ: если данный тэг не указан, то выводятся все поля.

#### Пример запроса 1

```bash
curl http://127.0.0.1:8081/ModelCalc -H "Content-Type: application/json" -X POST -d '{ "modelID" : "1234" ... }'
```

```JSON
{
    "modelID": "1003",
    "incommingParameters": [
        {
            "value": 5,
            "id": "P2014-07-22124737011-00835"
        },
        {
            "value": 2,
            "id": "P2014-07-22124822824-00590"
        }
    ],
    "outputParameters": [
        "P2014-10-17141000865-08165",
        "P2014-10-21124302106-09847"
    ],
    "service": {
        "outputFields": [
            "requiredExploredParameters",
            "timing"
        ]
    }
}
```

#### Пример ответа 1

Успешный ответ состоит из следующих объектов, если объект отсутствует в ответе, значит, что нет для него данных:

| Name of the object | Type | Consists of | Description |
|--------------------|------|-------------|-------------|
| requiredExploredParameters | Array | Объект Parameter | массив требуемых и найденных параметров |
| notRequiredExploredParameters | Array | Объект Parameter | массив найденных, но не требуемых параметров |
| requiredNotExploredParameters | Array | Значение Id | массив требуемых, но не найденных параметров |
| algorithm | Array | Объект AlgStep | массив правил, которые необходимы для вычисления требуемых параметров |
| Объект Parameter | Object | <ul> <li>id – значение Id</li> <li>value – значение ParameterValue</li> </ul> | Хранит идентификатор параметра и соответствующее ему значение |
| Значение Id | String | | Хранит уникальный идентификатор объекта в виде строки |
| Значение ParameterValue | String/Double/Boolean/Array | | Значение параметра, может быть один из 4ех типов, тип Array, пока не используется |
| Объект AlgStep | Object | <ul> <li>inputParameters – массив объектов modelRelParameter</li> <li>outputParameters – массив объектов modelRelParameter</li> <li>rule – объект Rule</li> <li>script – значение String</li></ul> | Хранит информацию об одном шаге алгоритма, входные параметры, найденные параметры, информация о правиле, использующийся скрипт |
| Объект modelRelParameter | Object | <ul> <li>modelParameterID – значение String, хранит уникальный идентификатор параметра из модели</li> <li>relationParameterID – значение String, хранит идентификатор переменной из скрипта</li> <li>value – значение String, хранит значение параметра</li> </ul> | Объект хранит информацию о соответствии параметра из модели к переменной из скрипта
| Объект Rule | Object | <ul> <li>id – Значение Id, уникальный идентификатор правила</li></ul> | Объект содержит всю требуемую информацию о правиле |
| Значение script | String | | Хранит скрипт, который требовался для расчета текущего шага алгоритма |

Пример ответа на запрос в случае успеха:

```JSON
{
    "requiredExploredParameters": [
        {
            "id": "P2014-10-17141000865-08165",
            "value": 11.582575694955839
        }
    ],
    "timing": {
        "requestOutputGeneration": 0,
        "requestParsing": 0,
        "requestProcessing": 1
    }
}
```

#### Пример запроса 2

Запрос 1 с добавление "algorithm" в "outputFields".

```bash
curl http://127.0.0.1:8081/ModelCalc -H "Content-Type: application/json" -X POST -d '{ "modelID" : "1234" ... }'
```

```JSON
{
    "modelID": "1003",
    "incommingParameters": [
        {
            "value": 5,
            "id": "P2014-07-22124737011-00835"
        },
        {
            "value": 2,
            "id": "P2014-07-22124822824-00590"
        }
    ],
    "outputParameters": [
        "P2014-10-17141000865-08165",
        "P2014-10-21124302106-09847"
    ],
    "service": {
        "outputFields": [
            "requiredExploredParameters",
            "timing",
            "algorithm"
        ]
    }
}
```

#### Пример ответа 2

```JSON
{
    "algorithm": [
        {
            "inputParameters": [
                {
                    "modelParameterID": "P2014-07-22124737011-00835",
                    "relationParameterID": "x1",
                    "value": "7.00"
                },
                {
                    "modelParameterID": "P2014-07-22124822824-00590",
                    "relationParameterID": "x2",
                    "value": "16.00"
                }
            ],
            "outputParameters": [
                {
                    "modelParameterID": "P2014-07-22124800380-08210",
                    "relationParameterID": "y",
                    "value": "21.00"
                }
            ],
            "rule": {
                "id": "R2014-11-26154536320-04673"
            },
            "script": "y=Math.sqrt(Math.pow(x1,2)-Math.pow(x2,2))"
        },
        {
            "inputParameters": [
                {
                    "modelParameterID": "P2014-07-22124737011-00835",
                    "relationParameterID": "BC",
                    "value": "10.00"
                },
                {
                    "modelParameterID": "P2014-07-22124800380-08210",
                    "relationParameterID": "AB",
                    "value": "10.00"
                },
                {
                    "modelParameterID": "P2014-07-22124822824-00590",
                    "relationParameterID": "AC",
                    "value": "1.00"
                }
            ],
            "outputParameters": [
                {
                    "modelParameterID": "P2014-10-17141000865-08165",
                    "relationParameterID": "P",
                    "value": "18.00"
                }
            ],
            "rule": {
                "id": "R2014-10-17142013771-03245"
            },
            "script": "P=BC+AB+AC"
        }
    ],
    "requiredExploredParameters": [
        {
            "id": "P2014-10-17141000865-08165",
            "value": 11.582575694955839
        }
    ],
    "timing": {
        "requestOutputGeneration": 0,
        "requestParsing": 0,
        "requestProcessing": 4
    }
}
```

Ответ, на запрос, в процессе вычисления которого сработало ограничение, состоит из следующих объектов (если объект отсутствует в ответе, значит, что нет для него данных):

```JSON
{
    "constraint": {
        "description": "Любая сторона треугольника меньше суммы двух других сторон и больше\nих разности",
        "id": "C2014-11-27163336351-00842"
    }
}
```

## Models

### Описание

| Property | Value |
|----------|-------|
| **Назначение** | Запрос на добавление новой модели в движок. |
| **Метод** | POST |
| **URL** | `/Models` |

### Структура JSON

#### Поля

| Property | Description |
|----------|-------------|
| **modelID** | ID модели. |
| **modelPoolSize** | Размер pool-а моделей, не обязательный параметр. Значение по умолчанию – 5, максимально возможное значение – 100. |
| **modelXML** | Модель в формате XML. |

#### Пример запроса

```bash
curl http://127.0.0.1:8081/Models -H "Content-Type: application/json" -X POST -d '{ "modelID" : "1234" ... }'
```

```JSON
{
    "modelID" : "1234",
    "modelPoolSize" : "5", 
    "modelXML" : "<?xml version=\"1.0\" encoding=\"UTF-8\"?><root><parameters> …" //модель в виде xml документа
}
```

#### Пример ответа

```JSON
{
    "modelID": "1234" //1234 - id модели
}
```

## ModelsCalcPackage
### Описание

| Property | Value |
|----------|-------|
| **Назначение** | Выполняет пакетную обработку запросов вычисления модели. |
| **Метод** | POST |
| **URL** | `/ModelsCalcPackage` |

### Структура JSON

#### Поля

| Property | Description |
|----------|-------------|
| **packages** | Массив, содержащий все пакеты запросов на вычисление модели. |

Каждый пакет содержит дополнительный тэг `packageID` – уникальный идентификатор пакета.
 - Каждый пакет содержит в себе все обязательные поля, перечисленные для API `/ModelsCalc`.
 - Каждый пакет может содержать в себе необязательные поля API `/ModelCals`.

#### Пример запроса

```bash
curl http://127.0.0.1:8081/ModelsCalcPackage -H "Content-Type: application/json" -X POST -d '{ "packages": [ ... }'
```

```JSON
{
  "packages": [
    {
      "packageID": "UID1",
      "modelID": "1003",
      "incommingParameters": [
        {
          "value": 5,
          "id": "P2014-07-22124737011-00835"
        },
        {
          "value": 2,
          "id": "P2014-07-22124822824-00590"
        }
      ],
      "outputParameters": [
        "P2014-10-17141000865-08165",
        "P2014-10-21124302106-09847"
      ],
      "service": {
        " outputFields": [
          "requiredExploredParameters",
          "timing"
        ]
      }
    },
    {
      "packageID": "UID2",
      "modelID": "1000",
      "incommingParameters": [
        {
          "value": 5,
          "id": "P2014-07-22124737011-00835"
        },
        {
          "value": 2,
          "id": "P2014-07-22124822824-00590"
        }
      ],
      "outputParameters": [
        "P2014-10-17141000865-08165",
        "P2014-10-21124302106-09847"
      ]
    }
  ]
}
```

#### Пример ответа

```JSON
[
  {
    "packageID": "UID1",
    "requiredExploredParameters": [
      {
        "id": "P2014-10-17141000865-08165",
        "value": 11.582575694955839
      }
    ],
    "timing": {
      "requestOutputGeneration": 0,
      "requestParsing": 0,
      "requestProcessing": 2
    }
  },
  {
    "ErrorDescription": "No model on Server",
    "ErrorID": 5101,
    "ErrorName": "NoModelOnServer",
    "packageID": "UID2"
  }
]
```

### Примечание

Обратите внимание, что ответ приходит для каждого пакета и ошибки обработки отдельного пакета не останавливают обработку всего запроса.

## ModelsParametersInfo
### Описание

| Property | Value |
|----------|-------|
| **Назначение** | Метод позволяет получить список параметров модели. |
| **Метод** | POST |
| **URL** | `/ModelsParametersInfo` |

### Структура JSON

#### Поля

| Property | Description |
|----------|-------------|
| **modelID** | Уникальный идентификатор модели на сервере. |
| **allowedFields** | Массив, содержащий поля параметров, которые будут содержаться в ответе. |

> Если поля `allowedFields` не будет в запросе, то каждый параметр будет содержать только поле `id`.

#### allowedFields (допустимые значения)

| Property | Description |
|----------|-------------|
| **type** | Тип значения параметра. |
| **description** | Описание параметра. |
| **defaultValue** | Значение по умолчанию (если есть). |

#### Пример запроса

```bash
curl http://127.0.0.1:8081/ModelsParametersInfo -H "Content-Type: application/json" -X POST -d '{ "modelID" : "1000" ... }'
```

```JSON
{
  "modelID": "1000",
  "allowedFields": [
    "defaultValue",
    "type"
  ]
}
```

#### Пример ответа

```JSON
{
  "parameters": [
    {
      "id": "P2014-10-23122506102-02475",
      "type": "double"
    },
    {
      "id": "P2014-10-22165422552-07550",
      "type": "double"
    },
    {
      "id": "P2014-10-21110010487-00928",
      "type": "double"
    },
    {
      "id": "P2014-10-29112331292-00028",
      "type": "double"
    },
    {
      "defaultValue": 90,
      "id": "P2014-10-21143417460-02700",
      "type": "double"
    }
  ]
}
```

## RelationParse
### Описание

| Property | Value |
|----------|-------|
| **Назначение** | Запрос парсинга отношений по ресурсу. |
| **Метод** | POST |
| **URL** | `/RelationParse` |

### Запрос для парсинга отношений типа if

#### Структура JSON

##### Поля

| Property | Description |
|----------|-------------|
| **relationType** | Тип отношения, для данного запроса тип - "ifthenelse". |
| **ifCode** | ЕСЛИ условного отношения, пример "X1==X2". |
| **thenCode** | ТО условного отношения, пример "Y=X2+4". |
| **elseCode** | ИНАЧЕ условного отношения, пример "Y=6". |

#### Пример запроса

```bash
curl http://127.0.0.1:8081/RelationParse -H "Content-Type: application/json" -X POST -d '{ "relationType" : "ifthenelse" ... }'
```

```JSON
{
  "relationType": "ifthenelse",
  "ifCode": "X1==X2",
  "thenCode": "Y=X2+4"
}
```

#### Пример ответа

```JSON
{
  "errorMessage": "Ok",
  "inputVars": [
    "X1",
    "X2"
  ],
  "outputVars": [
    "Y"
  ],
  "status": 200,
  "wholeIfThenElse": "if (X1 == X2) {Y=X1;} else {Y=6;}"
}
```

### Запрос для парcинга и проверки простых формулы

#### Структура JSON

| Property | Description |
|----------|-------------|
| **relationType** | Тип отношения, для данного запроса тип – "simple". |
| **simpleCode** | Простая математическая формула. |

#### Пример запроса

```bash
curl http://127.0.0.1:8081/RelationParse -H "Content-Type: application/json" -X POST -d '{ "relationType" : "simple" ... }'
```

```JSON
{
  "relationType": "simple",
  "simpleCode": "Y=X2+X1"
}
```

#### Пример ответа

```JSON
{
  "errorMessage": "Ok",
  "inputVars": [
    "X2",
    "X1"
  ],
  "outputVars": [
    "Y"
  ],
  "status": 200
}
```