# QOI — The “Quite OK Image Format” for fast, lossless image compression
Implementation of the QOI format using Go, along with a simple tool to visualise and analyze QOI's compression. 

See [QOI](https://qoiformat.org/) for an overview of what QOI is. The "quick rundown" is that QOI is an image format which:
1. Has a remarkably simple specification (it fits onto a single page).
2. Compresses images at high speed, locally and in a single pass.
3. Achieves a compression ratio close to that of PNG.

## Visualisation:
Since QOI compresses every pixel (or run of pixels, visualised with cyan) as an individual chunk, we are able to visualise QOI's compression by coloring the individual pixels according to which type of chunk was used. This library includes a function to automatise this process, and which also provides some basic statistical data on the results:

qoi_OP_RGB: Used to encode       6742 pixels (4.101%), using       20226 bytes (9.591%) total. Color: {255 0 0 255}
  
qoi_OP_RGBA: Used to encode          0 pixels (0.000%), using           0 bytes (0.000%) total. Color: {255 0 255 255}
 
qoi_OP_INDEX: Used to encode      38524 pixels (23.433%), using       38524 bytes (18.269%) total. Color: {0 255 0 255}

qoi_OP_DIFF: Used to encode      31311 pixels (19.046%), using       31311 bytes (14.848%) total. Color: {255 255 0 255}
 
qoi_OP_LUMA: Used to encode      52532 pixels (31.954%), using      105064 bytes (49.823%) total. Color: {0 0 255 255}
 
qoi_OP_RUN: Used to encode      35291 pixels (21.467%), using       15750 bytes (7.469%) total. Color: {0 255 255 255}
  
Overall compression ratio: 32.067368%


![a](https://raw.githubusercontent.com/HereComesTheMoon/QOI/master/testdata/qoi_test_images/cat.png) ![a](https://raw.githubusercontent.com/HereComesTheMoon/QOI/master/testdata/qoi_test_images/analysis/cat.png)
