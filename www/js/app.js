function updateElement(identifier, value) {
  var item = $(identifier),
      normalColour = 'rgb(178, 178, 178)',
      flashColour = '#FFF';

  if(item.text() == value) {
    return;
  }

  if(item.text() == '') {
    item.text(value);
    return;
  }

  if(window.timeouts[identifier]) {
    clearTimeout(window.timeouts[identifier]);
    delete(window.timeouts[identifier]);
  }

  item
    .text(value)
    .css('color', flashColour);

  window.timeouts[identifier] = setTimeout(function() {
    item.css('color', normalColour);
    delete(window.timeouts[identifier]);
  }, 1000);
}

function initPrices(prices) {
  var template = $('#template-price').html();

  for(var price in prices) {
    var item = template
                 .replace(/__NAME__/, price)
                 .replace(/__ID__/, 'price-' + price)
                 .replace(/__PRICE__/, prices[price].toFixed(2));
    $('.prices').append(item);
  }
}

function handlePrices(prices) {
  if(!$('.prices').children().length) {
    initPrices(prices);
    return;
  }

  for(var price in prices) {
    updateElement('.price-' + price, prices[price].toFixed(2));
  }
}

function initAssets(assets) {
  var templateVenue = $('#template-venue').html(),
      templateAsset = $('#template-asset').html();

  for(var venue in assets) {
    var $venue = $(templateVenue.replace(/__NAME__/, venue)),
        container = $venue.find('.venue-assets');

    for(var asset in assets[venue]) {
      var item = templateAsset
                   .replace(/__NAME__/, asset)
                   .replace(/__ID__/, 'asset-' + venue + '-' + asset)
                   .replace(/__QUANTITY__/, assets[venue][asset]);
      container.append(item);
    }

    $('.assets').append($venue);
  }
}

function handleAssets(assets) {
  if(!$('.assets').children().length) {
    initAssets(assets);
    return;
  }

  for(var venue in assets) {
    for(var asset in assets[venue]) {
      value = assets[venue][asset];
      if(asset == 'BTC') {
        value = value.toFixed(8);
      } else {
        value = value.toFixed(2);
      }

      updateElement('.asset-' + venue + '-' + asset, value);
    }
  }
}

function handlePayload(json) {
  if(json.assets) {
    handleAssets(json.assets);
  }

  if(json.prices) {
    handlePrices(json.prices);
  }

  var stats = ['exposure', 'cost', 'value', 'pnl', 'pnl_percentage', 'leverage_deribit', 'leverage_bybit'];
  for(var i in stats) {
    var stat = stats[i];

    if(!json.hasOwnProperty(stat)) {
      continue;
    }

    var value = json[stat];
    if(stat == 'exposure') {
      value = value.toFixed(8);
    } else if(['cost', 'value', 'pnl'].includes(stat)) {
      value = value.toLocaleString('en-US', { style: 'currency', currency: 'USD' });
    } else if(stat == 'pnl_percentage') {
      value = value.toFixed(2) + '%';
    } else if(stat.match(/^leverage_/)) {
      value = value.toFixed(2);
    }

    $('.stats').removeClass('d-none');
    updateElement('.stat-' + stat, value);
  }
}

$(function() {
  window.timeouts = {};
  var ws = new WebSocket('ws://' + window.location.host + '/ws');

  ws.onopen = function() {
    console.log('> onopen');
    var token = readCookie('auth_token') || window.prompt('Auth token');
    createCookie('auth_token', token, 365);
    ws.send(JSON.stringify({auth: token}));
  }

  ws.onclose = function() {
    console.log('> onclose');
  }

  ws.onerror = function(error) {
    console.log('> onerror');
    console.log(error);
  }

  ws.onmessage = function(message) {
    var json = JSON.parse(message.data);

    if(json.error) {
      window.alert('Authentication failed');
      eraseCookie('auth_token');
      return;
    }

    handlePayload(json);
  }
});
