# YAML 基础

## 步骤 01：注释和键值对
- 冒号后的空格是必需的，用于区分键和值
```yml
# 定义简单的键值对
name: kalyan
age: 23
city: Hyderabad
```

## 步骤 02：字典 / 映射
- 在一个项目后分组在一起的一组属性
- 字典下的所有项目都需要相同数量的空白空间
```yml
person:
  name: kalyan
  age: 23
  city: Hyderabad
```

## 步骤 03：数组 / 列表
- 破折号表示数组的一个元素
```yml
person: # 字典
  name: kalyan
  age: 23
  city: Hyderabad
  hobbies: # 列表  
    - cycling
    - cooking
  hobbies: [cycling, cooking]   # 使用不同表示法的列表  
```  

## 步骤 04：多个列表
- 破折号表示数组的一个元素
```yml
person: # 字典
  name: kalyan
  age: 23
  city: Hyderabad
  hobbies: # 列表  
    - cycling
    - cooking
  hobbies: [cycling, cooking]   # 使用不同表示法的列表  
  friends: # 
    - name: friend1
      age: 22
    - name: friend2
      age: 25            
```


## 步骤 05：Pod 模板示例供参考
```yml
apiVersion: v1 # 字符串
kind: Pod  # 字符串
metadata: # 字典
  name: myapp-pod
  labels: # 字典 
    app: myapp         
spec:
  containers: # 列表
    - name: myapp
      image: grissomsh/kubenginx:1.0.0
      ports:
        - containerPort: 80
          protocol: "TCP"
        - containerPort: 81
          protocol: "TCP"
```




