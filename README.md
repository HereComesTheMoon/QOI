# QOI — The “Quite OK Image Format” for fast, lossless image compression
Implementation of the QOI format using Go, along with a simple tool to visualise and analyze QOI's compression. 

See [QOI](https://qoiformat.org/) for an overview of what QOI is. The "quick rundown" is that QOI is an image format which:
1. Has a remarkably simple specification (it fits onto a single page).
2. Compresses images at high speed, locally and in a single pass.
3. Achieves a compression ratio close to that of PNG.

## Visualisation:
Since QOI compresses every pixel (or run of pixels, visualised with cyan) as an individual chunk, we are able to visualise QOI's compression by coloring the individual pixels according to which type of chunk was used. This library includes a function to automatise this process, and which also provides some basic statistical data on the results:

  qoi_OP_RGB was used to encode       1185 pixels ( 0.247%), using        3555 bytes ( 0.835%) total. Color: {255 0 0 255}
  
 qoi_OP_RGBA was used to encode      92749 pixels (19.323%), using      370996 bytes (87.150%) total. Color: {255 0 255 255}
 
qoi_OP_INDEX was used to encode       3493 pixels ( 0.728%), using        3493 bytes ( 0.821%) total. Color: {0 255 0 255}

 qoi_OP_DIFF was used to encode      13895 pixels ( 2.895%), using       13895 bytes ( 3.264%) total. Color: {255 255 0 255}
 
 qoi_OP_LUMA was used to encode       8870 pixels ( 1.848%), using       17740 bytes ( 4.167%) total. Color: {0 0 255 255}
 
  qoi_OP_RUN was used to encode     359808 pixels (74.960%), using       16018 bytes ( 3.763%) total. Color: {0 255 255 255}
  
Overall compression ratio: 22.171719%


![a](https://raw.githubusercontent.com/HereComesTheMoon/QOI/master/testdata/qoi_test_images/dice.png) 
![a](https://raw.githubusercontent.com/HereComesTheMoon/QOI/master/testdata/qoi_test_images/analysis/dice.png)
