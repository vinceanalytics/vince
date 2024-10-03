!function(){"use strict";var a=window.location,i=window.document,o=i.currentScript,l=o.getAttribute("data-api")||new URL(o.src).origin+"/api/event";function u(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(a.hostname)||"file:"===a.protocol)return u("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return u(null,t);try{if("true"===window.localStorage.plausible_ignore)return u("localStorage flag",t)}catch(e){}var n={},r=(n.n=e,n.u=t&&t.u?t.u:a.href,n.d=o.getAttribute("data-domain"),n.r=i.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),n.h=1,new XMLHttpRequest);r.open("POST",l,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(n)),r.onreadystatechange=function(){4===r.readyState&&t&&t.callback&&t.callback({status:r.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var n=0;n<t.length;n++)e.apply(this,t[n]);function s(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var c=1;function r(e){var t,n;if("auxclick"!==e.type||e.button===c)return(t=function(e){for(;e&&(void 0===e.tagName||!s(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],!function e(t,n){if(!t||d<n)return!1;if(v(t))return!0;return e(t.parentNode,n+1)}(t,0)&&(n=t)&&n.href&&n.host&&n.host!==a.host?p(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):void 0}function p(e,t,n){var r,a=!1;function i(){a||(a=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(r={props:n.props},plausible(n.name,r)):(r={props:n.props,callback:i},plausible(n.name,r),setTimeout(i,5e3),e.preventDefault())}function f(e){var e=v(e)?e:e&&e.parentNode,t={name:null,props:{}},n=e&&e.classList;if(n)for(var r=0;r<n.length;r++){var a,i=n.item(r).match(/plausible-event-(.+)(=|--)(.+)/);i&&(a=i[1],i=i[3].replace(/\+/g," "),"name"==a.toLowerCase()?t.name=i:t.props[a]=i)}return t}i.addEventListener("click",r),i.addEventListener("auxclick",r);var d=3;function m(e){if("auxclick"!==e.type||e.button===c){for(var t,n,r,a,i=e.target,o=0;o<=d&&i;o++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;s(i)&&(t=i),v(i)&&(n=i),i=i.parentNode}n&&(a=f(n),t?(a.props.url=t.href,p(e,t,a)):((e={}).props=a.props,plausible(a.name,e)))}}function v(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}i.addEventListener("submit",function(e){var t,n=e.target,r=f(n);function a(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),e={props:r.props,callback:a},plausible(r.name,e))}),i.addEventListener("click",m),i.addEventListener("auxclick",m)}();