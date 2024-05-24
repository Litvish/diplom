const readline = require('readline');
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

let gameRunning = true;
let playerPosition = 1; // 0 - левая дорога, 1 - центральная дорога, 2 - правая дорога
let incomingCars = [false, false, false]; // Позиции встречных машин

const roads = ['левая дорога', 'центральная дорога', 'правая дорога'];

function displayRoads() {
  console.clear();
  roads.forEach((road, index) => {
    let roadString = road;
    if (incomingCars[index]) {
      roadString += ' <--- встречная машина!';
    }
    if (index === playerPosition) {
      roadString = `[${roadString}]`;
    }
    console.log(roadString);
  });
}

function movePlayer(direction) {
  if (direction === 'left' && playerPosition > 0) {
    playerPosition--;
  } else if (direction === 'right' && playerPosition < 2) {
    playerPosition++;
  }
  displayRoads();
}

function spawnIncomingCars() {
  incomingCars = incomingCars.map(() => Math.random() < 0.3); // 30% шанс появления машины на каждой дороге
}

function checkCollision() {
  if (incomingCars[playerPosition]) {
    console.log('Столкновение! Игра окончена.');
    gameRunning = false;
  } else {
    console.log('Вы успешно избежали столкновения!');
  }
}

function gameLoop() {
  if (!gameRunning) {
    rl.close();
    return;
  }

  spawnIncomingCars();
  displayRoads();

  rl.question('Выберите дорогу (left, right): ', (answer) => {
    movePlayer(answer);
    checkCollision();
    gameLoop();
  });
}

// Начало игры
gameLoop();