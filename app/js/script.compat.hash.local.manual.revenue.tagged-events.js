!function(){"use strict";var e,t,i=window.location,u=window.document,o=u.getElementById("plausible"),l=o.getAttribute("data-api")||(e=(e=o).src.split("/"),t=e[0],e=e[2],t+"//"+e+"/api/event");function n(e,t){try{if("true"===window.localStorage.plausible_ignore)return n=t,(a="localStorage flag")&&console.warn("Ignoring Event: "+a),void(n&&n.callback&&n.callback())}catch(e){}var n,a={},r=(a.n=e,a.u=t&&t.u?t.u:i.href,a.d=o.getAttribute("data-domain"),a.r=u.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),a.h=1,new XMLHttpRequest);r.open("POST",l,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(a)),r.onreadystatechange=function(){4===r.readyState&&t&&t.callback&&t.callback({status:r.status})}}var a=window.plausible&&window.plausible.q||[];window.plausible=n;for(var r=0;r<a.length;r++)n.apply(this,a[r]);function s(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var p=1;function c(e){"auxclick"===e.type&&e.button!==p||((e=function(e){for(;e&&(void 0===e.tagName||!s(e)||!e.href);)e=e.parentNode;return e}(e.target))&&e.href&&e.href.split("?")[0],function e(t,n){if(!t||d<n)return!1;if(g(t))return!0;return e(t.parentNode,n+1)}(e,0))}function f(e,t,n){var a,r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((a={props:n.props}).revenue=n.revenue,plausible(n.name,a)):((a={props:n.props,callback:i}).revenue=n.revenue,plausible(n.name,a),setTimeout(i,5e3),e.preventDefault())}function v(e){var e=g(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},n=e&&e.classList;if(n)for(var a=0;a<n.length;a++){var r,i,u=n.item(a),o=u.match(/plausible-event-(.+)(=|--)(.+)/),o=(o&&(r=o[1],i=o[3].replace(/\+/g," "),"name"==r.toLowerCase()?t.name=i:t.props[r]=i),u.match(/plausible-revenue-(.+)(=|--)(.+)/));o&&(r=o[1],i=o[3],t.revenue[r]=i)}return t}u.addEventListener("click",c),u.addEventListener("auxclick",c);var d=3;function m(e){if("auxclick"!==e.type||e.button===p){for(var t,n,a,r,i=e.target,u=0;u<=d&&i;u++){if((a=i)&&a.tagName&&"form"===a.tagName.toLowerCase())return;s(i)&&(t=i),g(i)&&(n=i),i=i.parentNode}n&&(r=v(n),t?(r.props.url=t.href,f(e,t,r)):((e={}).props=r.props,e.revenue=r.revenue,plausible(r.name,e)))}}function g(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}u.addEventListener("submit",function(e){var t,n=e.target,a=v(n);function r(){t||(t=!0,n.submit())}a.name&&(e.preventDefault(),t=!1,setTimeout(r,5e3),(e={props:a.props,callback:r}).revenue=a.revenue,plausible(a.name,e))}),u.addEventListener("click",m),u.addEventListener("auxclick",m)}();