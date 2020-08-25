function addValue() {
  //alert("ready go");
  //document.getElementById('demo').innerHTML = Date();
  var num = document.getElementById('demo').innerHTML;
  num = parseInt(num); 
  num = num + 1;
  //alert(num);
  document.getElementById('demo').innerHTML = num;
}

function changeValue(){
  alert("fires");
  var str = document.getElementById('inputbox').innerText;
  alert(str);
  var length = str.length;
  document.getElementById('wordCount').innerHTML = length;
}

let myButton = document.getElementById('myButton'); 
myButton.onclick = addValue;

let userInputBox = document.getElementById('inputbox');
userInputBox.onchange = changeValue;
