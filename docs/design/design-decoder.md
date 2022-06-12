# Shifu Decode Module
- [Shifu Decode Module](#shifu-decode-module)
  - [Design Goals and Non-goals](#design-goals-and-non-goals)
    - [Design Goals](#design-goals)
    - [Design Non-goals](#design-non-goals)
  - [Message Formats](#message-formats)
    - [Number](#number)
    - [String](#string)
  - [Decode process](#decode-process)
## Design Goals and Non-goals

**Why do we need this decode module** 

Currently Shifu cannot decode and analyze the message in the device response; instead it can only present the original raw data. This could affect the user experience as the user needs to purchase or perform additional work to decode the message.

Therefore Shifu should have the ability to decode the device resopnse message so that anyone can understand the message once Shifu shows it.

**Fundamental feature**

Decode module should take and process the message Shifu gets from device, and present the user-friendly message with high readability.

**Solution** 

Our biggest and only chanllenge is the huge variety of device types. We must assume that each type of device will have different protocols and different message formates.

Just like what we did for Shifu to make it compatible with different protocols, we have 2 methods to let Shifu understand the messages:

1. For standardized, widely-used and popular format, Shifu will have a pre-defined processor;
2. for new, private and unknwon format, Shifu will allow user to define the method to decode the message.

### Design Goals

Shifu decode module should have the following features:
1. Running on every deviceShifu - so deviceShifu itself does the decoding of the message;
2. every deviceShifu should only have the format the device is associated, which is defined in the configuration file.

### Design Non-goals
This design does not include the process the physical signals. For now Shifu should only focus on digital messages.

## Message Formats

Shifu should support the message format including but not limited to:

- character

  - number

    - binary

    - decimal 

    - hexadecimal

  - string

    - JSON 

    - protobuf 

    - XML 

    - formats defined by user

    - no format

- physical signals

### Number

For numbers (like values in PLC memory), we need to have 2 parameters:

- bits: represents Shifu should read from which bit to which bit, interval closed at left and opened at right

- endian: the read order of the binary number

We will need user to define the 2 parameters above, a parameter for message type, a parameter for API name, and a parameter for the meaning. Here is a simple example of the configuation:

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

When deviceShifu receives the message, it will know the 0th bit is used for describing whether the device is open.
  

### String

**Standardized formats**

We just need to include the already-defined libraries:

```
instruction: open 

responseMsgType: json 
```

**Formats defined by user** 

We need user to provide the dictionary for delimiters and keys:

```
Instruction: open 

responseMsgType: formattedString 

dictionaryDilimiter: ',' 

dictionary： 

    age 

    name 

    gender 
```

Therefore when Shifu receives a string "75,Trump,Male" we shall know it is about a 75-year old male named Trump.

**String with no formats**

Shifu should show the string with its original raw format:

```
Instruction: open 

messageType: string 
```

We should keep expanding the compatible formats. We will have a list of compatible formats in the doc next.
  
## Decode process
Here is the typical decode process：
[![decoder-flow](/img/decoder-flow.svg)](/img/decoder-flow.svg)    
