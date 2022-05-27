# Shifu 解码模块（Decode Module） 
- [Shifu 解码模块（Decode Module）](#shifu-解码模块decode-module)
  - [设计目标与非目标](#设计目标与非目标)
    - [设计目标](#设计目标)
    - [设计非目标](#设计非目标)
  - [信息格式](#信息格式)
    - [纯数字](#纯数字)
    - [字符串](#字符串)
  - [解码流程](#解码流程)
## 设计目标与非目标

**引入解码模块必要性** 

当前Shifu无法对设备发出的信息进行解码和分析，只能忠实地照原样呈现原始信息。这会给用户的使用带来很大不便 —— 用户需要额外再编写或购买解码服务对原始信息进行处理。 

Shifu应当拥有这种功能，使用户不需要借助外力就可以第一时间理解设备发出的信息。 

**必须实现的功能**

简而言之，解码模块需要对设备发出的信息进行处理，并给用户呈现出具备实际意义和完全可读性的信息。 

**挑战和对策** 

最大而唯一的挑战就是无法预估的繁多的设备种类。 

我们必须假设，每种设备使用的通信协议、信息格式都完全不同。 

因此，与Shifu兼容不同通信协议的方法类似，我们需要两手准备： 

1. 对通用的、流行的信息格式，Shifu应当有默认库进行兼容。 

2. 对新的、未知的信息格式，允许用户进行自定义，告诉Shifu如何解码。 

### 设计目标

Shifu的解码模块应当满足下列要求： 

1. 运行于每个deviceShifu之上，即deviceShifu负责对信息进行解码。 

2. 每个deviceShifu上运行的解码模块应当只包含这个deviceShifu对应的信息格式，信息格式应由用户在配置文件中说明。 


### 设计非目标
本设计并不包括对于设备物理信号的解析和处理。当前，Shifu应该专注于数字化信息。  

## 信息格式 

Shifu需要支持的信息格式包括但不限于： 

- 字符 

  - 纯数字 

  - 二进制 

  - 十进制 

  - 十六进制 

- 字符串 

  - JSON 

  - protobuf 

  - XML 

  - 其他自定义格式 

  - 无格式 

- 原始物理量 

当前，我们需要把注意力放在字符这一类信息里。 

### 纯数字 

对于纯数字（如PLC内存值），我们需要确定两个参数： 

- bit数量（bits）：表示从第几个bit读到第几个bit，区间左闭右开 

- 序（endian）：字节序，通俗理解就是阅读顺序 

也就是说，我们需要用户来定义以上两个参数、一个表示类型的参数、一个表示API的参数和一个表示意义的参数。一个简单的配置例子如下： 
```
Instruction: open 

responseMsgType: binary 

endian: little 

 

fields:  

    field:isOpen 

    bits: 0 - 1 

    field:timeMs 

    bits: 1 - 5 

    type: int 
```

这样，deviceShifu在收到信息之后，就会知道第零位的数字表达的是是否open了。 

  

### 字符串

**通用格式的字符串**

我们只需要集成对应的库即可。 
```
instruction: open 

responseMsgType: json 
```

**自定义格式的字符串** 

我们应当需要用户提供字典，配置格式为提供分隔符和字典关键字。 

```
Instruction: open 

responseMsgType: formattedString 

dictionaryDilimiter: ',' 

dictionary： 

    age 

    name 

    gender 
```

这样，如果我们收到一个字符串"75,Trump,Male"，Shifu就会知道这是在说一个名叫Trump的75岁男性了。 

  

**无格式字符串**

当前，Shifu应当按照原样把无格式字符串呈现出来。 

```
Instruction: open 

messageType: string 
```

我们需要不断扩充Shifu兼容的通用流行格式。当前我们需要维护一个Shifu兼容格式的列表，加入文档。 

  
## 解码流程
一个典型的解码流程如下图：
[![decoder-flow](/img/decoder-flow.svg)](/img/decoder-flow.svg)    
