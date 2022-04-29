import torch
import torch.nn as nn
from PIL import Image  # 导入图片处理工具
import PIL.ImageOps
import numpy as np
from torchvision import transforms
# import cv2
import matplotlib.pyplot as plt

class CNN(nn.Module):
    def __init__(self):
        super(CNN, self).__init__()
        self.conv1 = nn.Sequential(  # input shape (1, 28, 28)
            nn.Conv2d(
                in_channels=1,  # 输入通道数
                out_channels=16,  # 输出通道数
                kernel_size=5,  # 卷积核大小
                stride=1,  #卷积步数
                padding=2,  # 如果想要 con2d 出来的图片长宽没有变化, 
                            # padding=(kernel_size-1)/2 当 stride=1
            ),  # output shape (16, 28, 28)
            nn.ReLU(),  # activation
            nn.MaxPool2d(kernel_size=2))  # 在 2x2 空间里向下采样, output shape (16, 14, 14) )
        self.conv2 = nn.Sequential(  # input shape (16, 14, 14)
            nn.Conv2d(16, 32, 5, 1, 2),  # output shape (32, 14, 14)
            nn.ReLU(),  # activation
            nn.MaxPool2d(2))  # output shape (32, 7, 7) )
        self.out = nn.Linear(32 * 7 * 7, 10)  # 全连接层，0-9一共10个类
    # 前向反馈
    def forward(self, x):
        x = self.conv1(x)
        x = self.conv2(x)
        # print(x.size(3))
        x = x.view(x.size(0), -1)  # 展平多维的卷积图成 (batch_size, 32 * 7 * 7)
        output = self.out(x)
        return output

file_name = '3.png'  # 导入自己的图片
img = Image.open(file_name)
# plt.imshow(img)
# plt.show()
img = img.convert('L')
# plt.imshow(img)
# plt.show()
img = PIL.ImageOps.invert(img)
# plt.imshow(img)
# plt.show()
train_transform = transforms.Compose([
       transforms.Grayscale(),
         transforms.Resize((28, 28)),
         transforms.ToTensor(),
 ])
img = train_transform(img)
print(img.size())
print(img[0 , 1, 1])
img = torch.unsqueeze(img, dim=0)
print(img[0, 0 , 1, 1])


model = CNN()
index_to_class = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9']

with torch.no_grad():
    y = model(img)
    output = torch.squeeze(y)
    predict = torch.softmax(output, dim=0)
    print(predict)
    predict_cla = torch.argmax(predict).numpy()
    print(predict_cla)
print(index_to_class[predict_cla], predict[predict_cla].numpy())